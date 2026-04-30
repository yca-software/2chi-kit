export interface User {
  id: string;
  createdAt: string;

  firstName: string;
  lastName: string;
  language: string;
  avatarURL: string;

  email: string;
  emailVerifiedAt: string | null;
  googleId: string | null;

  termsAcceptedAt: string;
  termsVersion: string;
}

export interface AdminAccess {
  userId: string;
  createdAt: string;
}

export interface UserRefreshToken {
  id: string;
  userId: string;

  createdAt: string;
  expiresAt: string;
  revokedAt: string | null;

  ip: string;
  userAgent: string;
}

export interface UserPasswordResetToken {
  id: string;
  userId: string;

  createdAt: string;
  expiresAt: string;
  usedAt: string | null;
}

export interface UserEmailVerificationToken {
  id: string;
  userId: string;

  createdAt: string;
  expiresAt: string;
  usedAt: string | null;
}
