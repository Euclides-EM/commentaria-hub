import { MetadataService } from "@hub-api";
import { useQuery } from "@tanstack/react-query";

export function usePseudonyms() {
  return useQuery({
    queryKey: ["metadata", "pseudonyms"],
    queryFn: () => MetadataService.getPseudonyms(),
    staleTime: Infinity,
    refetchOnWindowFocus: false,
  });
}
