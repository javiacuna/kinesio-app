import { QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider } from "@/features/auth/AuthProvider";
import { queryClient } from "./queryClient";

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>{children}</AuthProvider>
    </QueryClientProvider>
  );
}
