package parcel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrParcelNotFound         = errors.New("parcel not found")
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrUnauthorizedAccess     = errors.New("unauthorized resource access")
	ErrDuplicateIdempotency   = errors.New("duplicate request idempotency match")
)

type ParcelFilter struct {
	TenantID       uuid.UUID
	SenderID       *uuid.UUID // Scoped if role is CUSTOMER
	Status         *ParcelStatus
	ServiceType    *ServiceType
	OriginBranchID *uuid.UUID
	SearchQuery    string
	Limit          int
	Offset         int
}

type Repository interface {
	CreateParcel(ctx context.Context, p *Parcel) (*Parcel, error)
	GetParcelByID(ctx context.Context, tenantID, id uuid.UUID) (*Parcel, error)
	GetParcelByTracking(ctx context.Context, trackingNumber string) (*Parcel, error)
	GetParcelByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*Parcel, error)
	ListParcels(ctx context.Context, filter ParcelFilter) ([]*Parcel, int, error)
	UpdateParcelStatus(ctx context.Context, tenantID, parcelID uuid.UUID, newStatus ParcelStatus, actorID *uuid.UUID, branchID *uuid.UUID, notes string, locationLat, locationLon *float64) (*Parcel, error)
	
	// Address Book
	CreateAddress(ctx context.Context, a *Address) (*Address, error)
	GetAddressByID(ctx context.Context, customerID, id uuid.UUID) (*Address, error)
	ListCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*Address, error)
	DeleteAddress(ctx context.Context, customerID, id uuid.UUID) error

	// Custody & Audit
	AddCustodyEvent(ctx context.Context, e *CustodyEvent) error
	GetCustodyTimeline(ctx context.Context, parcelID uuid.UUID) ([]*CustodyEvent, error)
}

type sqlRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) CreateParcel(ctx context.Context, p *Parcel) (*Parcel, error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt

	query := `
		INSERT INTO parcels (
			id, tenant_id, tracking_number, sender_id, origin_branch_id, destination_branch_id,
			service_type, status, sender_name, sender_phone, sender_email, sender_address,
			sender_city, sender_state, sender_postal_code, sender_location,
			recipient_name, recipient_phone, recipient_email, recipient_address,
			recipient_city, recipient_state, recipient_postal_code, recipient_location,
			weight_kg, length_cm, width_cm, height_cm, declared_value, currency,
			is_fragile, special_instructions, shipping_cost, payment_status,
			qr_token, idempotency_key, estimated_delivery_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14, $15, CASE WHEN $16::float8 IS NOT NULL AND $17::float8 IS NOT NULL THEN ST_SetSRID(ST_MakePoint($16, $17), 4326) ELSE NULL END,
			$18, $19, $20, $21,
			$22, $23, $24, CASE WHEN $25::float8 IS NOT NULL AND $26::float8 IS NOT NULL THEN ST_SetSRID(ST_MakePoint($25, $26), 4326) ELSE NULL END,
			$27, $28, $29, $30, $31, $32,
			$33, $34, $35, $36,
			$37, $38, $39, $40, $41
		)
		RETURNING id, created_at, updated_at;
	`

	var sLon, sLat, rLon, rLat *float64
	if p.SenderLongitude != nil && p.SenderLatitude != nil {
		sLon, sLat = p.SenderLongitude, p.SenderLatitude
	}
	if p.RecipientLongitude != nil && p.RecipientLatitude != nil {
		rLon, rLat = p.RecipientLongitude, p.RecipientLatitude
	}

	err := r.db.QueryRowContext(ctx, query,
		p.ID, p.TenantID, p.TrackingNumber, p.SenderID, p.OriginBranchID, p.DestinationBranchID,
		p.ServiceType, p.Status, p.SenderName, p.SenderPhone, p.SenderEmail, p.SenderAddress,
		p.SenderCity, p.SenderState, p.SenderPostalCode, sLon, sLat,
		p.RecipientName, p.RecipientPhone, p.RecipientEmail, p.RecipientAddress,
		p.RecipientCity, p.RecipientState, p.RecipientPostalCode, rLon, rLat,
		p.WeightKg, p.LengthCm, p.WidthCm, p.HeightCm, p.DeclaredValue, p.Currency,
		p.IsFragile, p.SpecialInstructions, p.ShippingCost, p.PaymentStatus,
		p.QRToken, p.IdempotencyKey, p.EstimatedDeliveryAt, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert parcel: %w", err)
	}

	// Record initial custody intake event
	_ = r.AddCustodyEvent(ctx, &CustodyEvent{
		ID:             uuid.New(),
		ParcelID:       p.ID,
		TenantID:       p.TenantID,
		ActorID:        &p.SenderID,
		BranchID:       p.OriginBranchID,
		EventType:      "CREATED",
		PreviousStatus: "",
		NewStatus:      StatusCreated,
		LocationName:   fmt.Sprintf("%s, %s", p.SenderCity, p.SenderPostalCode),
		Notes:          "Parcel consignment booked via portal",
		RecordedAt:     p.CreatedAt,
	})

	return p, nil
}

func (r *sqlRepository) GetParcelByID(ctx context.Context, tenantID, id uuid.UUID) (*Parcel, error) {
	query := `
		SELECT 
			id, tenant_id, tracking_number, sender_id, origin_branch_id, destination_branch_id, current_branch_id,
			service_type, status, sender_name, sender_phone, sender_email, sender_address,
			sender_city, sender_state, sender_postal_code, ST_X(sender_location), ST_Y(sender_location),
			recipient_name, recipient_phone, recipient_email, recipient_address,
			recipient_city, recipient_state, recipient_postal_code, ST_X(recipient_location), ST_Y(recipient_location),
			weight_kg, length_cm, width_cm, height_cm, declared_value, currency,
			is_fragile, special_instructions, shipping_cost, payment_status,
			qr_token, idempotency_key, estimated_delivery_at, delivered_at, created_at, updated_at
		FROM parcels
		WHERE id = $1 AND tenant_id = $2;
	`
	p := &Parcel{}
	var originBranch, destBranch, currBranch *uuid.UUID
	var sLon, sLat, rLon, rLat *float64
	var idempKey *string

	err := r.db.QueryRowContext(ctx, query, id, tenantID).Scan(
		&p.ID, &p.TenantID, &p.TrackingNumber, &p.SenderID, &originBranch, &destBranch, &currBranch,
		&p.ServiceType, &p.Status, &p.SenderName, &p.SenderPhone, &p.SenderEmail, &p.SenderAddress,
		&p.SenderCity, &p.SenderState, &p.SenderPostalCode, &sLon, &sLat,
		&p.RecipientName, &p.RecipientPhone, &p.RecipientEmail, &p.RecipientAddress,
		&p.RecipientCity, &p.RecipientState, &p.RecipientPostalCode, &rLon, &rLat,
		&p.WeightKg, &p.LengthCm, &p.WidthCm, &p.HeightCm, &p.DeclaredValue, &p.Currency,
		&p.IsFragile, &p.SpecialInstructions, &p.ShippingCost, &p.PaymentStatus,
		&p.QRToken, &idempKey, &p.EstimatedDeliveryAt, &p.DeliveredAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrParcelNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get parcel: %w", err)
	}

	p.OriginBranchID = originBranch
	p.DestinationBranchID = destBranch
	p.CurrentBranchID = currBranch
	p.SenderLongitude = sLon
	p.SenderLatitude = sLat
	p.RecipientLongitude = rLon
	p.RecipientLatitude = rLat
	if idempKey != nil {
		p.IdempotencyKey = *idempKey
	}
	return p, nil
}

func (r *sqlRepository) GetParcelByTracking(ctx context.Context, trackingNumber string) (*Parcel, error) {
	query := `
		SELECT 
			id, tenant_id, tracking_number, sender_id, origin_branch_id, destination_branch_id, current_branch_id,
			service_type, status, sender_name, sender_phone, sender_email, sender_address,
			sender_city, sender_state, sender_postal_code, ST_X(sender_location), ST_Y(sender_location),
			recipient_name, recipient_phone, recipient_email, recipient_address,
			recipient_city, recipient_state, recipient_postal_code, ST_X(recipient_location), ST_Y(recipient_location),
			weight_kg, length_cm, width_cm, height_cm, declared_value, currency,
			is_fragile, special_instructions, shipping_cost, payment_status,
			qr_token, idempotency_key, estimated_delivery_at, delivered_at, created_at, updated_at
		FROM parcels
		WHERE UPPER(tracking_number) = UPPER($1);
	`
	p := &Parcel{}
	var originBranch, destBranch, currBranch *uuid.UUID
	var sLon, sLat, rLon, rLat *float64
	var idempKey *string

	err := r.db.QueryRowContext(ctx, query, trackingNumber).Scan(
		&p.ID, &p.TenantID, &p.TrackingNumber, &p.SenderID, &originBranch, &destBranch, &currBranch,
		&p.ServiceType, &p.Status, &p.SenderName, &p.SenderPhone, &p.SenderEmail, &p.SenderAddress,
		&p.SenderCity, &p.SenderState, &p.SenderPostalCode, &sLon, &sLat,
		&p.RecipientName, &p.RecipientPhone, &p.RecipientEmail, &p.RecipientAddress,
		&p.RecipientCity, &p.RecipientState, &p.RecipientPostalCode, &rLon, &rLat,
		&p.WeightKg, &p.LengthCm, &p.WidthCm, &p.HeightCm, &p.DeclaredValue, &p.Currency,
		&p.IsFragile, &p.SpecialInstructions, &p.ShippingCost, &p.PaymentStatus,
		&p.QRToken, &idempKey, &p.EstimatedDeliveryAt, &p.DeliveredAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrParcelNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get parcel by tracking: %w", err)
	}

	p.OriginBranchID = originBranch
	p.DestinationBranchID = destBranch
	p.CurrentBranchID = currBranch
	p.SenderLongitude = sLon
	p.SenderLatitude = sLat
	p.RecipientLongitude = rLon
	p.RecipientLatitude = rLat
	if idempKey != nil {
		p.IdempotencyKey = *idempKey
	}
	return p, nil
}

func (r *sqlRepository) GetParcelByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*Parcel, error) {
	query := `
		SELECT 
			id, tenant_id, tracking_number, sender_id, origin_branch_id, destination_branch_id, current_branch_id,
			service_type, status, sender_name, sender_phone, sender_email, sender_address,
			sender_city, sender_state, sender_postal_code, ST_X(sender_location), ST_Y(sender_location),
			recipient_name, recipient_phone, recipient_email, recipient_address,
			recipient_city, recipient_state, recipient_postal_code, ST_X(recipient_location), ST_Y(recipient_location),
			weight_kg, length_cm, width_cm, height_cm, declared_value, currency,
			is_fragile, special_instructions, shipping_cost, payment_status,
			qr_token, idempotency_key, estimated_delivery_at, delivered_at, created_at, updated_at
		FROM parcels
		WHERE tenant_id = $1 AND idempotency_key = $2;
	`
	p := &Parcel{}
	var originBranch, destBranch, currBranch *uuid.UUID
	var sLon, sLat, rLon, rLat *float64
	var idempKey *string

	err := r.db.QueryRowContext(ctx, query, tenantID, key).Scan(
		&p.ID, &p.TenantID, &p.TrackingNumber, &p.SenderID, &originBranch, &destBranch, &currBranch,
		&p.ServiceType, &p.Status, &p.SenderName, &p.SenderPhone, &p.SenderEmail, &p.SenderAddress,
		&p.SenderCity, &p.SenderState, &p.SenderPostalCode, &sLon, &sLat,
		&p.RecipientName, &p.RecipientPhone, &p.RecipientEmail, &p.RecipientAddress,
		&p.RecipientCity, &p.RecipientState, &p.RecipientPostalCode, &rLon, &rLat,
		&p.WeightKg, &p.LengthCm, &p.WidthCm, &p.HeightCm, &p.DeclaredValue, &p.Currency,
		&p.IsFragile, &p.SpecialInstructions, &p.ShippingCost, &p.PaymentStatus,
		&p.QRToken, &idempKey, &p.EstimatedDeliveryAt, &p.DeliveredAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Not found is normal for fresh idempotency checks
	}
	if err != nil {
		return nil, fmt.Errorf("idempotency check query failed: %w", err)
	}

	p.OriginBranchID = originBranch
	p.DestinationBranchID = destBranch
	p.CurrentBranchID = currBranch
	p.SenderLongitude = sLon
	p.SenderLatitude = sLat
	p.RecipientLongitude = rLon
	p.RecipientLatitude = rLat
	if idempKey != nil {
		p.IdempotencyKey = *idempKey
	}
	return p, nil
}

func (r *sqlRepository) ListParcels(ctx context.Context, filter ParcelFilter) ([]*Parcel, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	baseWhere := "WHERE tenant_id = $1"
	args := []interface{}{filter.TenantID}
	argIdx := 2

	if filter.SenderID != nil {
		baseWhere += fmt.Sprintf(" AND sender_id = $%d", argIdx)
		args = append(args, *filter.SenderID)
		argIdx++
	}

	if filter.Status != nil {
		baseWhere += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.ServiceType != nil {
		baseWhere += fmt.Sprintf(" AND service_type = $%d", argIdx)
		args = append(args, string(*filter.ServiceType))
		argIdx++
	}

	if filter.SearchQuery != "" {
		baseWhere += fmt.Sprintf(" AND (tracking_number ILIKE $%d OR recipient_name ILIKE $%d OR recipient_phone ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+filter.SearchQuery+"%")
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM parcels " + baseWhere
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count parcels: %w", err)
	}

	// Select rows
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, tenant_id, tracking_number, sender_id, origin_branch_id, destination_branch_id, current_branch_id,
			service_type, status, sender_name, sender_phone, sender_email, sender_address,
			sender_city, sender_state, sender_postal_code, ST_X(sender_location), ST_Y(sender_location),
			recipient_name, recipient_phone, recipient_email, recipient_address,
			recipient_city, recipient_state, recipient_postal_code, ST_X(recipient_location), ST_Y(recipient_location),
			weight_kg, length_cm, width_cm, height_cm, declared_value, currency,
			is_fragile, special_instructions, shipping_cost, payment_status,
			qr_token, idempotency_key, estimated_delivery_at, delivered_at, created_at, updated_at
		FROM parcels
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d;
	`, baseWhere, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query parcels: %w", err)
	}
	defer rows.Close()

	var result []*Parcel
	for rows.Next() {
		p := &Parcel{}
		var originBranch, destBranch, currBranch *uuid.UUID
		var sLon, sLat, rLon, rLat *float64
		var idempKey *string

		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.TrackingNumber, &p.SenderID, &originBranch, &destBranch, &currBranch,
			&p.ServiceType, &p.Status, &p.SenderName, &p.SenderPhone, &p.SenderEmail, &p.SenderAddress,
			&p.SenderCity, &p.SenderState, &p.SenderPostalCode, &sLon, &sLat,
			&p.RecipientName, &p.RecipientPhone, &p.RecipientEmail, &p.RecipientAddress,
			&p.RecipientCity, &p.RecipientState, &p.RecipientPostalCode, &rLon, &rLat,
			&p.WeightKg, &p.LengthCm, &p.WidthCm, &p.HeightCm, &p.DeclaredValue, &p.Currency,
			&p.IsFragile, &p.SpecialInstructions, &p.ShippingCost, &p.PaymentStatus,
			&p.QRToken, &idempKey, &p.EstimatedDeliveryAt, &p.DeliveredAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan parcel row: %w", err)
		}

		p.OriginBranchID = originBranch
		p.DestinationBranchID = destBranch
		p.CurrentBranchID = currBranch
		p.SenderLongitude = sLon
		p.SenderLatitude = sLat
		p.RecipientLongitude = rLon
		p.RecipientLatitude = rLat
		if idempKey != nil {
			p.IdempotencyKey = *idempKey
		}
		result = append(result, p)
	}

	return result, total, nil
}

func (r *sqlRepository) UpdateParcelStatus(
	ctx context.Context,
	tenantID, parcelID uuid.UUID,
	newStatus ParcelStatus,
	actorID *uuid.UUID,
	branchID *uuid.UUID,
	notes string,
	locationLat, locationLon *float64,
) (*Parcel, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Fetch current parcel status with row lock
	var currentStatus ParcelStatus
	err = tx.QueryRowContext(ctx, "SELECT status FROM parcels WHERE id = $1 AND tenant_id = $2 FOR UPDATE", parcelID, tenantID).Scan(&currentStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrParcelNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current status: %w", err)
	}

	// 2. Enforce strict table-driven state machine transition
	if !CanTransition(currentStatus, newStatus) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStateTransition, currentStatus, newStatus)
	}

	now := time.Now().UTC()
	var deliveredAt *time.Time
	if newStatus == StatusDelivered {
		deliveredAt = &now
	}

	// 3. Update parcel table
	updateQuery := `
		UPDATE parcels
		SET status = $1, current_branch_id = COALESCE($2, current_branch_id),
		    delivered_at = COALESCE($3, delivered_at), updated_at = $4
		WHERE id = $5 AND tenant_id = $6;
	`
	_, err = tx.ExecContext(ctx, updateQuery, newStatus, branchID, deliveredAt, now, parcelID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to update parcel: %w", err)
	}

	// 4. Log immutable custody event inside same transaction
	custodyQuery := `
		INSERT INTO custody_events (
			id, parcel_id, tenant_id, actor_id, branch_id, event_type,
			previous_status, new_status, location, notes, recorded_at
		) VALUES (
			$1, $2, $3, $4, $5, 'STATUS_CHANGE',
			$6, $7, CASE WHEN $8::float8 IS NOT NULL AND $9::float8 IS NOT NULL THEN ST_SetSRID(ST_MakePoint($8, $9), 4326) ELSE NULL END,
			$10, $11
		);
	`
	custodyID := uuid.New()
	_, err = tx.ExecContext(ctx, custodyQuery,
		custodyID, parcelID, tenantID, actorID, branchID,
		currentStatus, newStatus, locationLon, locationLat,
		notes, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record custody event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit status update: %w", err)
	}

	return r.GetParcelByID(ctx, tenantID, parcelID)
}

func (r *sqlRepository) CreateAddress(ctx context.Context, a *Address) (*Address, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	a.CreatedAt = time.Now().UTC()
	a.UpdatedAt = a.CreatedAt

	query := `
		INSERT INTO addresses (
			id, customer_id, tenant_id, label, contact_name, contact_phone, contact_email,
			street_line1, street_line2, city, state, postal_code, country,
			location, is_default_pickup, is_default_delivery, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			CASE WHEN $14::float8 IS NOT NULL AND $15::float8 IS NOT NULL THEN ST_SetSRID(ST_MakePoint($14, $15), 4326) ELSE NULL END,
			$16, $17, $18, $19
		) RETURNING id, created_at, updated_at;
	`
	err := r.db.QueryRowContext(ctx, query,
		a.ID, a.CustomerID, a.TenantID, a.Label, a.ContactName, a.ContactPhone, a.ContactEmail,
		a.StreetLine1, a.StreetLine2, a.City, a.State, a.PostalCode, a.Country,
		a.Longitude, a.Latitude, a.IsDefaultPickup, a.IsDefaultDelivery, a.CreatedAt, a.UpdatedAt,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert address: %w", err)
	}
	return a, nil
}

func (r *sqlRepository) GetAddressByID(ctx context.Context, customerID, id uuid.UUID) (*Address, error) {
	query := `
		SELECT 
			id, customer_id, tenant_id, label, contact_name, contact_phone, contact_email,
			street_line1, street_line2, city, state, postal_code, country,
			ST_X(location), ST_Y(location), is_default_pickup, is_default_delivery, created_at, updated_at
		FROM addresses
		WHERE id = $1 AND customer_id = $2;
	`
	a := &Address{}
	err := r.db.QueryRowContext(ctx, query, id, customerID).Scan(
		&a.ID, &a.CustomerID, &a.TenantID, &a.Label, &a.ContactName, &a.ContactPhone, &a.ContactEmail,
		&a.StreetLine1, &a.StreetLine2, &a.City, &a.State, &a.PostalCode, &a.Country,
		&a.Longitude, &a.Latitude, &a.IsDefaultPickup, &a.IsDefaultDelivery, &a.CreatedAt, &a.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("address not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	return a, nil
}

func (r *sqlRepository) ListCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*Address, error) {
	query := `
		SELECT 
			id, customer_id, tenant_id, label, contact_name, contact_phone, contact_email,
			street_line1, street_line2, city, state, postal_code, country,
			ST_X(location), ST_Y(location), is_default_pickup, is_default_delivery, created_at, updated_at
		FROM addresses
		WHERE customer_id = $1
		ORDER BY is_default_pickup DESC, created_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list addresses: %w", err)
	}
	defer rows.Close()

	var list []*Address
	for rows.Next() {
		a := &Address{}
		if err := rows.Scan(
			&a.ID, &a.CustomerID, &a.TenantID, &a.Label, &a.ContactName, &a.ContactPhone, &a.ContactEmail,
			&a.StreetLine1, &a.StreetLine2, &a.City, &a.State, &a.PostalCode, &a.Country,
			&a.Longitude, &a.Latitude, &a.IsDefaultPickup, &a.IsDefaultDelivery, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan address: %w", err)
		}
		list = append(list, a)
	}
	return list, nil
}

func (r *sqlRepository) DeleteAddress(ctx context.Context, customerID, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM addresses WHERE id = $1 AND customer_id = $2", id, customerID)
	if err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("address not found")
	}
	return nil
}

func (r *sqlRepository) AddCustodyEvent(ctx context.Context, e *CustodyEvent) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.RecordedAt.IsZero() {
		e.RecordedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO custody_events (
			id, parcel_id, tenant_id, actor_id, branch_id, event_type,
			previous_status, new_status, location, location_name, device_info, notes, recorded_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, CASE WHEN $9::float8 IS NOT NULL AND $10::float8 IS NOT NULL THEN ST_SetSRID(ST_MakePoint($9, $10), 4326) ELSE NULL END,
			$11, $12, $13, $14
		);
	`
	_, err := r.db.ExecContext(ctx, query,
		e.ID, e.ParcelID, e.TenantID, e.ActorID, e.BranchID, e.EventType,
		e.PreviousStatus, e.NewStatus, e.Longitude, e.Latitude,
		e.LocationName, e.DeviceInfo, e.Notes, e.RecordedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert custody event: %w", err)
	}
	return nil
}

func (r *sqlRepository) GetCustodyTimeline(ctx context.Context, parcelID uuid.UUID) ([]*CustodyEvent, error) {
	query := `
		SELECT 
			c.id, c.parcel_id, c.tenant_id, c.actor_id, COALESCE(u.first_name || ' ' || u.last_name, 'System') as actor_name,
			c.branch_id, COALESCE(b.name, '') as branch_name, c.event_type,
			COALESCE(c.previous_status, ''), COALESCE(c.new_status, ''),
			ST_Y(c.location) as lat, ST_X(c.location) as lon,
			COALESCE(c.location_name, ''), COALESCE(c.device_info, ''), COALESCE(c.notes, ''), c.recorded_at
		FROM custody_events c
		LEFT JOIN users u ON c.actor_id = u.id
		LEFT JOIN branches b ON c.branch_id = b.id
		WHERE c.parcel_id = $1
		ORDER BY c.recorded_at ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, parcelID)
	if err != nil {
		return nil, fmt.Errorf("failed to query custody timeline: %w", err)
	}
	defer rows.Close()

	var events []*CustodyEvent
	for rows.Next() {
		e := &CustodyEvent{}
		var actorID, branchID *uuid.UUID
		var prev, next string
		var lat, lon *float64

		if err := rows.Scan(
			&e.ID, &e.ParcelID, &e.TenantID, &actorID, &e.ActorName,
			&branchID, &e.BranchName, &e.EventType,
			&prev, &next, &lat, &lon,
			&e.LocationName, &e.DeviceInfo, &e.Notes, &e.RecordedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan custody row: %w", err)
		}
		e.ActorID = actorID
		e.BranchID = branchID
		e.PreviousStatus = ParcelStatus(prev)
		e.NewStatus = ParcelStatus(next)
		e.Latitude = lat
		e.Longitude = lon
		events = append(events, e)
	}
	return events, nil
}
