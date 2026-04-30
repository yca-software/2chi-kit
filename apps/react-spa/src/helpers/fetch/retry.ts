import { RequestError } from ".";

export const evaluateRetry = (failureCount: number, err: any): boolean => {
  const { status, retry } = err as RequestError;

  if (!retry) {
    return false;
  }

  if (status >= 400 && status < 500 && status !== 401) {
    return false;
  }

  return failureCount < 3;
};
