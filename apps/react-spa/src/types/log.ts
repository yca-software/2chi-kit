export interface AuditLog {
  id: string;
  createdAt: string;

  organizationId: string;

  actorId: string;
  actorInfo: string;
  impersonatedById: string | null;
  impersonatedByEmail: string;

  action: string;
  resourceType: string;
  resourceId: string;
  resourceName: string | null;

  data: Record<string, any> | null;
}
