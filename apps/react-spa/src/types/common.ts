export interface PaginatedListResponse<T> {
  items: T[];
  hasNext: boolean;
}

export interface APIError {
  error: Error;
  errorCode: string;
  message: string;
  extra?: unknown;
}

export interface MutationError {
  status: number;
  error: APIError;
  retry?: boolean;
}

export interface MutationCallbacks<TResponse = unknown> {
  onSuccess?: (data: TResponse) => void;
  onError?: (error: MutationError) => void;
}

export interface JWTAccessTokenPermissionData {
  organizationId: string;
  roleId: string;
  permissions: string[];
}

export interface AccessInfoFromToken {
  sub: string;
  email: string;
  exp: number;
  iat: number;
  permissions: JWTAccessTokenPermissionData[];
  isAdmin: boolean;
}
