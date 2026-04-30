import { useAPI } from "@/helpers";
import { useMutation } from "@tanstack/react-query";
import type { MutationError } from "@/types";

export interface SupportRequest {
  subject: string;
  message: string;
  pageUrl?: string;
}

export const useSupportMutation = (callbacks?: {
  onSuccess?: () => void;
  onError?: (err: MutationError) => void;
}) => {
  const fetchWrapper = useAPI();
  return useMutation<void, MutationError, SupportRequest>({
    mutationFn: (body: SupportRequest) => {
      const payload = {
        ...body,
        pageUrl:
          typeof window !== "undefined" ? window.location.href : undefined,
      };
      return fetchWrapper({
        endpoint: "support",
        method: "POST",
        body: payload,
      }) as Promise<void>;
    },
    onSuccess: callbacks?.onSuccess,
    onError: callbacks?.onError,
  });
};
