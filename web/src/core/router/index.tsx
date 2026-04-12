import { lazy, Suspense } from "react";
import {
  BrowserRouter,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { ProtectedRoute } from "@/components/protected-route";

const LoginPage = lazy(() =>
  import("@/features/auth/pages").then((m) => ({ default: m.LoginPage }))
);
const TradingPage = lazy(() =>
  import("@/features/trading/pages").then((m) => ({ default: m.TradingPage }))
);
const OrderDetailPage = lazy(() =>
  import("@/features/trading/pages").then((m) => ({ default: m.OrderDetailPage }))
);
const WalletPage = lazy(() =>
  import("@/features/wallet/pages").then((m) => ({ default: m.WalletPage }))
);

function PageLoader() {
  return (
    <div className="flex h-screen items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
    </div>
  );
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/trade"
            element={
              <ProtectedRoute>
                <TradingPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/orders/:id"
            element={
              <ProtectedRoute>
                <OrderDetailPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/wallet"
            element={
              <ProtectedRoute>
                <WalletPage />
              </ProtectedRoute>
            }
          />
          <Route path="*" element={<Navigate to="/trade" replace />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  );
}
