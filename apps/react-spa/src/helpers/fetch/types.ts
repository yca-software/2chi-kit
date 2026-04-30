export interface RequestConfig {
  endpoint: string;
  method: string;
  body?: any;
  multipart?: boolean;
}

export interface RequestError {
  error?: any;
  status: number;
  retry?: boolean;
}
