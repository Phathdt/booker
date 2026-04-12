import { Toaster } from "@/components/ui/sonner";
import { DataProvider, AuthProvider } from "@/core/providers";
import { AppRouter } from "@/core/router";
import type { ComponentType, ReactNode } from "react";

type ProviderComponent = ComponentType<{ children: ReactNode }>;

function composeProviders(...providers: ProviderComponent[]) {
  return providers.reduceRight<ProviderComponent>(
    (Acc, Provider) =>
      ({ children }) => (
        <Provider>
          <Acc>{children}</Acc>
        </Provider>
      ),
    ({ children }) => <>{children}</>
  );
}

const Providers = composeProviders(DataProvider, AuthProvider);

export default function App() {
  return (
    <Providers>
      <AppRouter />
      <Toaster richColors position="top-right" />
    </Providers>
  );
}
