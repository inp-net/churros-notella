//@ts-nocheck

import { connect, JetStreamManager, NatsConnection, StringCodec } from "nats"
import {
  STREAM_NAME,
  SUBJECT_NAME,
  type Message,
  Event,
} from "../typescript/index.js"

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

function randomMessage(): Message {
  const events = [
    Event.GodchildRequest,
    Event.NewPost,
    Event.NewTicket,
  ] as const
  const event = events[Math.floor(Math.random() * events.length)]
  const id = Math.random().toString(36).substring(7)
  const send_at = new Date()
  send_at.setSeconds(send_at.getSeconds() + Math.floor(Math.random() * 10))
  return {
    event,
    id,
    send_at,
    clear_schedule_for:
      Math.random() > 0.5
        ? []
        : Math.random() > 0.5
        ? [Event.GodchildRequest, event]
        : [event],
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
    await js.publish(subject, sc.encode(JSON.stringify(randomMessage())))
    console.log(`Sent message: ${message}`)
    await new Promise((resolve) => setTimeout(resolve, delayMs))
  }
}

async function main() {
  // Connect to the NATS server
  const nc = await connect({ servers: "localhost:4222" })
  console.log("Connected to NATS")

  // Create a JetStream manager to manage streams
  const jsm = await nc.jetstreamManager()

  // Ensure the stream exists
  await setupStream(jsm, STREAM_NAME, SUBJECT_NAME)

  // Publish messages at intervals
  const messageCount = 1000
  const delayMs = 10 // 1 second delay between messages
  await publishMessages(nc, SUBJECT_NAME, messageCount, delayMs)

  console.log("Finished sending messages.")
  await nc.close()
}

main().catch((err) => {
  console.error(`Error in sender: ${err.message}`)
})
