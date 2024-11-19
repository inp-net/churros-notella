import * as z from "zod";


export const ConfigurationSchema = z.object({
    "APP_PACKAGE_ID": z.string(),
    "CONTACT_EMAIL": z.string(),
    "DATABASE_URL": z.string(),
    "FIREBASE_SERVICE_ACCOUNT": z.string(),
    "HEALTH_CHECK_PORT": z.number(),
    "NATS_URL": z.string(),
    "PUBLIC_VAPID_KEY": z.string(),
    "REDIS_URL": z.string(),
    "STARTUP_SCHEDULE_RESTORATION": z.string(),
    "VAPID_PRIVATE_KEY": z.string(),
});
export type Configuration = z.infer<typeof ConfigurationSchema>;
