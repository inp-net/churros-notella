export interface Configuration {
    APP_PACKAGE_ID:               string;
    CONTACT_EMAIL:                string;
    DATABASE_URL:                 string;
    FIREBASE_SERVICE_ACCOUNT:     string;
    HEALTH_CHECK_PORT:            number;
    NATS_URL:                     string;
    PUBLIC_VAPID_KEY:             string;
    REDIS_URL:                    string;
    STARTUP_SCHEDULE_RESTORATION: string;
    VAPID_PRIVATE_KEY:            string;
}
