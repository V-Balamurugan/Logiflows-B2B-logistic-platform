# LogiFlows — Database Schema Foundation

## 1. Spatial & Foundation Extensions

LogiFlows requires PostGIS for spatial polygon bounding and coordinates representation:
- `uuid-ossp`: For generating cryptographically strong UUIDv4 identifiers.
- `postgis`: For geometry and geography point/polygon data types.

## 2. Core Relational Entities (Master Schema Roadmap)

1. `tenants`: Multi-tenant companies isolating all logistics entities.
2. `users`: Identity and authentication records.
3. `roles` & `permissions`: Granular RBAC tables.
4. `branches`: Physical sorting hubs, distribution centers, and serviceability polygons.
5. `employees`: Couriers, sorting staff, dispatchers, and branch managers.
6. `vehicles`: Fleet inventory with payload and cubic volume capacities.
7. `customers` & `customer_addresses`: Senders and recipients with verified geocoordinates.
8. `parcels`: Master parcel business records, QR tokens, and lifecycle state machines.
9. `custody_events`: Tamper-evident chain-of-custody transfer logs.
10. `routes` & `route_stops`: Scheduled sequence of dispatch stops.
11. `tracking_points`: Courier telemetry location trail with PostGIS spatial points.
12. `delay_predictions`: Advisory AI risk logs.
13. `proof_of_delivery`: Digital signatures, recipient verification, and storage links.
