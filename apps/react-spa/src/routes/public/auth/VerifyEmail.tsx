import { useVerifyEmailMutation } from "@/api";
import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { useTranslationNamespace } from "@/helpers";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Button,
  Link,
  Alert,
  AlertDescription,
} from "@yca-software/design-system";
import { CheckCircle, XCircle, ArrowLeft, Loader2 } from "lucide-react";

export const VerifyEmail = () => {
  const { t, isLoading } = useTranslationNamespace(["auth", "common"]);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const [isVerified, setIsVerified] = useState(false);
  const [isVerifying, setIsVerifying] = useState(false);

  const {
    mutate: verifyEmail,
    isPending,
    error,
  } = useVerifyEmailMutation({
    onSuccess: () => {
      setIsVerified(true);
      setIsVerifying(false);
      setTimeout(() => {
        navigate("/");
      }, 3000);
    },
    onError: () => {
      setIsVerifying(false);
    },
  });

  useEffect(() => {
    if (token && !isVerifying && !isVerified) {
      setIsVerifying(true);
      verifyEmail({ token });
    }
  }, [token, isVerifying, isVerified, verifyEmail]);

  if (isLoading) {
    return null;
  }

  if (!token) {
    return (
      <Card className="w-full">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10">
            <XCircle className="h-6 w-6 text-destructive" />
          </div>
          <CardTitle className="text-2xl">
            {t("auth:verifyEmail.invalidTokenTitle")}
          </CardTitle>
          <CardDescription>
            {t("auth:verifyEmail.invalidTokenDescription")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" asChild className="w-full">
            <Link to="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              {t("auth:verifyEmail.backToSignIn")}
            </Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (isPending || isVerifying) {
    return (
      <Card className="w-full">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center">
            <Loader2 className="h-6 w-6 animate-spin text-primary" />
          </div>
          <CardTitle className="text-2xl">
            {t("auth:verifyEmail.verifyingTitle")}
          </CardTitle>
          <CardDescription>
            {t("auth:verifyEmail.verifyingDescription")}
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  if (isVerified) {
    return (
      <Card className="w-full">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
            <CheckCircle className="h-6 w-6 text-primary" />
          </div>
          <CardTitle className="text-2xl">
            {t("auth:verifyEmail.successTitle")}
          </CardTitle>
          <CardDescription>
            {t("auth:verifyEmail.successDescription")}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Button variant="outline" asChild className="w-full">
            <Link to="/">
              <ArrowLeft className="mr-2 h-4 w-4" />
              {t("auth:verifyEmail.backToSignIn")}
            </Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader className="space-y-1">
        <CardTitle className="text-2xl">
          {t("auth:verifyEmail.title")}
        </CardTitle>
        <CardDescription>{t("auth:verifyEmail.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        {error && (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>
              {error.error?.message || t("defaultError", { ns: "common" })}
            </AlertDescription>
          </Alert>
        )}
        <Button variant="outline" asChild className="w-full">
          <Link to="/">
            <ArrowLeft className="mr-2 h-4 w-4" />
            {t("auth:verifyEmail.backToSignIn")}
          </Link>
        </Button>
      </CardContent>
    </Card>
  );
};
