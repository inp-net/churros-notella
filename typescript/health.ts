import * as z from "zod";


export const HealthResponseSchema = z.object({
    "churros_db": z.boolean(),
    "firebase": z.boolean(),
    "nats": z.boolean(),
    "redis": z.boolean(),
});
export type HealthResponse = z.infer<typeof HealthResponseSchema>;
