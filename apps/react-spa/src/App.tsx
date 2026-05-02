import { Helmet } from "react-helmet-async";
import {
  ErrorBoundary,
  PaddleProvider,
  PageviewTracker,
  CookieBanner,
  LegalAcceptanceGate,
} from "./components";
import { BrowserRouter, Navigate, Route, Routes } from "react-router";
import { Root } from "./routes/Root";
import {
  LegalLayout,
  TermsOfService,
  CookiePolicy,
  AuthLayout,
  SignIn,
  SignUp,
  ForgotPassword,
  ResetPassword,
  VerifyEmail,
  GoogleOAuthCallback,
} from "./routes/public";
import {
  AppLayout,
  Dashboard,
  GeneralSettings,
  ProtectedRoute,
  RequireOrganization,
  Settings,
  TeamsSettings,
  SettingsLayout,
  RolesSettings,
  ApiKeysSettings,
  AuditLogSettings,
  MembersSettings,
} from "./routes/private";
import {
  AdminLayout,
  AdminOrganizationDetail,
  AdminOrganizations,
  AdminProtectedRoute,
  AdminUserDetail,
  AdminUsers,
} from "./routes/admin";
import { AdminDashboard } from "./routes/admin/AdminDashboard";

export const App = () => {
  return (
    <ErrorBoundary>
      <Helmet>
        <title>Example App Title</title>
      </Helmet>
      <BrowserRouter>
        <PaddleProvider>
          <PageviewTracker />
          <CookieBanner />
          <Routes>
            <Route element={<Root />}>
              <Route element={<LegalLayout />}>
                <Route path="/terms-of-service" element={<TermsOfService />} />
                <Route path="/cookie-policy" element={<CookiePolicy />} />
              </Route>

              <Route element={<AuthLayout />}>
                <Route path="/" element={<SignIn />} />
                <Route path="/signup" element={<SignUp />} />
                <Route path="/forgot-password" element={<ForgotPassword />} />
                <Route path="/reset-password" element={<ResetPassword />} />
                <Route path="/verify-email" element={<VerifyEmail />} />
                <Route path="/oauth/google" element={<GoogleOAuthCallback />} />
              </Route>

              <Route
                element={
                  <ProtectedRoute>
                    <LegalAcceptanceGate>
                      <AppLayout />
                    </LegalAcceptanceGate>
                  </ProtectedRoute>
                }
              >
                <Route path="/dashboard/:orgId" element={<Dashboard />} />
                <Route path="/dashboard" element={<Dashboard />} />

                <Route element={<RequireOrganization />}>
                  <Route path="/settings/:orgId" element={<SettingsLayout />}>
                    <Route index element={<Settings />} />
                    <Route path="general" element={<GeneralSettings />} />
                    <Route path="roles" element={<RolesSettings />} />
                    <Route path="teams" element={<TeamsSettings />} />
                    <Route path="api-keys" element={<ApiKeysSettings />} />
                    <Route path="audit-log" element={<AuditLogSettings />} />
                    <Route path="members" element={<MembersSettings />} />
                  </Route>
                  <Route path="/settings" element={<Settings />} />
                </Route>
              </Route>

              <Route element={<AdminProtectedRoute />}>
                <Route path="admin" element={<AdminLayout />}>
                  <Route index element={<AdminDashboard />} />
                  <Route path="users" element={<AdminUsers />} />
                  <Route path="users/:userId" element={<AdminUserDetail />} />
                  <Route
                    path="organizations"
                    element={<AdminOrganizations />}
                  />
                  <Route
                    path="organizations/archived/:orgId"
                    element={<AdminOrganizationDetail />}
                  />
                  <Route
                    path="organizations/:orgId"
                    element={<AdminOrganizationDetail />}
                  />
                </Route>
              </Route>

              <Route path="*" element={<Navigate to="/" replace />} />
            </Route>
          </Routes>
        </PaddleProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
};
