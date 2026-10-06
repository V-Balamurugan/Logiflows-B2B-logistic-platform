export type Role =
  | 'PLATFORM_ADMIN'
  | 'TENANT'
  | 'TENANT_ADMIN'
  | 'EMPLOYEE'
  | 'CUSTOMER';

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  role: Role;
  is_active: boolean;
  tenant_id?: string;
  created_at: string;
  updated_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  token_type: string;
}

export interface AuthResponse {
  user: User;
  tokens: TokenPair;
  permissions: string[];
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface RegisterPayload {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  phone?: string;
  role: 'CUSTOMER' | 'TENANT';
  company_name?: string;
}
