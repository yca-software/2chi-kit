import { useAPI } from "@/helpers";
import { useQuery } from "@tanstack/react-query";

export interface LocationAutocompleteResponse {
  predictions?: Array<{
    placeId: string;
    description: string;
    structuredFormatting?: {
      mainText: string;
      secondaryText: string;
    };
  }>;
}

export function useLocationAutocompleteQuery(input: string, enabled: boolean) {
  const fetchWrapper = useAPI();
  return useQuery({
    queryKey: ["location", "autocomplete", input],
    queryFn: () =>
      fetchWrapper({
        endpoint: `location/autocomplete?input=${encodeURIComponent(input)}`,
        method: "GET",
      }) as Promise<LocationAutocompleteResponse>,
    enabled: enabled && input.trim().length > 0,
  });
}
