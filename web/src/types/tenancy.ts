export type BranchType = 'SORTING_HUB' | 'DISTRIBUTION_CENTER' | 'LOCAL_OFFICE';
export type BranchStatus = 'ACTIVE' | 'INACTIVE' | 'MAINTENANCE';

export interface GeoPoint {
  latitude: number;
  longitude: number;
}

export interface OperatingHours {
  open: string;
  close: string;
  timezone: string;
}

export interface Branch {
  id: string;
  tenant_id: string;
  code: string;
  name: string;
  branch_type: BranchType;
  address: string;
  city: string;
  state: string;
  postal_code: string;
  country: string;
  contact_phone: string;
  contact_email: string;
  operating_hours: OperatingHours;
  daily_capacity: number;
  status: BranchStatus;
  location: GeoPoint;
  service_area: any;
  created_at: string;
  updated_at: string;
}

export interface BranchSummary {
  id: string;
  code: string;
  name: string;
  branch_type: BranchType;
  address: string;
  city: string;
  contact_phone: string;
  location: GeoPoint;
  status: BranchStatus;
}

export interface BranchCreatePayload {
  code: string;
  name: string;
  branch_type: BranchType;
  address: string;
  city: string;
  state: string;
  postal_code: string;
  country?: string;
  contact_phone: string;
  contact_email: string;
  daily_capacity: number;
  status: BranchStatus;
  location: GeoPoint;
  service_area: any;
}

export interface ServiceabilityCheckRequest {
  tenant_id: string;
  latitude: number;
  longitude: number;
  address?: string;
}

export interface ServiceabilityCheckResponse {
  is_serviceable: boolean;
  status: string;
  message: string;
  matched_branch?: BranchSummary;
  nearest_branch?: BranchSummary;
  distance_meters: number;
  estimated_eta?: string;
}

export interface GeoJSONFeature {
  type: string;
  geometry: {
    type: string;
    coordinates: number[][][] | number[][][][];
  };
  properties: {
    branch_id: string;
    code: string;
    name: string;
    branch_type: BranchType;
    status: BranchStatus;
  };
}

export interface GeoJSONFeatureCollection {
  type: string;
  features: GeoJSONFeature[];
}

export interface TenantEnterprise {
  id: string;
  name: string;
  code: string;
  slug: string;
  tier: string;
  quota_parcels_per_day: number;
  quota_branches: number;
  branding: {
    primary_color: string;
    accent_color: string;
    portal_name: string;
    logo_url?: string;
  };
  config: {
    timezone: string;
    currency: string;
    auto_assign_delivery: boolean;
  };
  is_active: boolean;
  created_at: string;
  updated_at: string;
}
