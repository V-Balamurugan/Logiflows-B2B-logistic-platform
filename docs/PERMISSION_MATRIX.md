# LogiFlows Authoritative Permission Matrix

All permissions are centrally declared, deny-by-default, and strictly enforced on the Go backend at the HTTP middleware and service layer.

| Permission Key | Description | PLATFORM_ADMIN | TENANT | TENANT_ADMIN | EMPLOYEE | CUSTOMER |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: |
| `tenant.read` | View tenant profiles/metadata | Yes | Yes (Own) | Yes (Own) | No | No |
| `tenant.create` | Provision enterprise tenants | Yes | No | No | No | No |
| `tenant.update` | Modify tenant settings/branding | Yes | No | No | No | No |
| `tenant.disable` | Suspend tenant operations | Yes | No | No | No | No |
| `platform_admin.create`| Provision platform administrator | Yes | No | No | No | No |
| `tenant_admin.create` | Provision tenant administrator | Yes | Yes | No | No | No |
| `employee.create` | Provision courier or branch staff | No | Yes | Yes | No | No |
| `employee.read` | View employee roster and status | Yes | Yes | Yes | No | No |
| `employee.update` | Update staff details / shifts | No | Yes | Yes | No | No |
| `customer.create` | Onboard customer accounts | No | Yes | Yes | No | Self |
| `customer.read` | Access customer profiles | No | Yes | Yes | No | Own |
| `customer.update` | Update customer address/details | No | Yes | Yes | No | Own |
| `parcel.create` | Book a new consignment | No | Yes | Yes | No | Yes (Own) |
| `parcel.read` | Read parcel specifications | Yes | Yes | Yes | Assigned | Own |
| `parcel.update` | Update parcel dimensions/status | No | Yes | Yes | No | No |
| `parcel.cancel` | Void/cancel parcel booking | No | Yes | Yes | No | Own (Pending) |
| `parcel.track` | Access tracking events | Yes | Yes | Yes | Yes | Yes |
| `parcel.qr.generate` | Generate signed QR token | No | Yes | Yes | Yes | No |
| `parcel.qr.download` | Download PDF label with QR | No | Yes | Yes | Yes | No |
| `custody.read` | Inspect custody handover chain | No | Yes | Yes | Yes | No |
| `custody.create` | Log intake, scan, or transfer | No | No | Yes | Yes | No |
| `custody.correct` | Administrative correction | No | No | No | No | No |
| `delivery.assign` | Dispatch parcel to driver/hub | No | Yes | Yes | No | No |
| `delivery.update` | Update delivery state (POD) | No | Yes | Yes | Yes | No |
| `branch.manage` | Create/update hub boundaries | Yes | Yes | Yes | No | No |
| `vehicle.manage` | Manage fleet vehicle assets | Yes | Yes | Yes | No | No |
| `vehicle.read` | View fleet inventory and specs | Yes | Yes | Yes | Yes | No |
| `route.review` | Inspect delivery routes & ETAs | No | Yes | Yes | Yes | No |
| `report.read` | Export analytics & operational KPIs| Yes | Yes | Yes | No | No |
| `audit.read` | Inspect immutable audit logs | Yes | Yes | No | No | No |
| `system.settings` | Modify global platform parameters | Yes | No | No | No | No |
