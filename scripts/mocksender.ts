import { connect, StringCodec, JetStreamManager, NatsConnection } from "nats"

async function setupStream(
  jetStreamManager: JetStreamManager,
  streamName: string,
  subject: string
) {
  try {
    // Try to add the stream (will not recreate if it exists)
    await jetStreamManager.streams.add({
      name: streamName,
      subjects: [subject],
    })
    console.log(`Stream '${streamName}' created or already exists.`)
  } catch (err) {
    console.error(`Error setting up stream: ${err.message}`)
  }
}

async function publishMessages(
  nc: NatsConnection,
  subject: string,
  messageCount: number,
  delayMs: number
) {
  const js = nc.jetstream()
  const sc = StringCodec()

  for (let i = 1; i <= messageCount; i++) {
    const message = `Mock Order #${i}`
    await js.publish(subject, sc.encode(message))
    console.log(`Sent message: ${message}`)
    await new Promise((resolve) => setTimeout(resolve, delayMs))
  }
}

async function main() {
  const streamName = "notella:stream"
  const subject = "notella:notification"

  // Connect to the NATS server
  const nc = await connect({ servers: "localhost:4222" })
  console.log("Connected to NATS")

  // Create a JetStream manager to manage streams
  const jsm = await nc.jetstreamManager()

  // Ensure the stream exists
  await setupStream(jsm, streamName, subject)

  // Publish messages at intervals
  const messageCount = 20
  const delayMs = 1000 // 1 second delay between messages
  await publishMessages(nc, subject, messageCount, delayMs)

  console.log("Finished sending messages.")
  await nc.close()
}

main().catch((err) => {
  console.error(`Error in sender: ${err.message}`)
})
