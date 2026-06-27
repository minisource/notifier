export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface PaginationParams {
  page?: number;
  pageSize?: number;
}

export type SortDirection = 'asc' | 'desc';

export interface ApiError {
  code: string;
  message: string;
  details?: unknown;
  requestId?: string;
  status?: number;
}

export interface ApiResponse<T> {
  data: T;
  requestId?: string;
}

export interface DateRange {
  from: string;
  to: string;
}

export interface SortParams {
  field: string;
  direction: SortDirection;
}
