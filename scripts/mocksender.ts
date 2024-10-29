//@ts-nocheck

import { connect, JetStreamManager, NatsConnection, StringCodec } from 'nats';
import { STREAM_NAME, SUBJECT_NAME } from '../constants.js';
import { Message } from '../types.js';

async function setupStream(
  jetStreamManager: JetStreamManager,
  streamName: string,
  subject: string,
) {
  try {
    // Try to add the stream (will not recreate if it exists)
    await jetStreamManager.streams.add({
      name: streamName,
      subjects: [subject],
    });
    console.log(`Stream '${streamName}' created or already exists.`);
  } catch (err) {
    console.error(`Error setting up stream: ${err.message}`);
  }
}

enum Event {
  CommentReply = 'comment_reply',
  GodchildRequest = 'godchild_request',
  NewComment = 'new_comment',
  NewPost = 'new_post',
  NewTicket = 'new_ticket',
}

function randomMessage(): Message {
  const events = [
    Event.CommentReply,
    Event.GodchildRequest,
    Event.NewComment,
    Event.NewPost,
    Event.NewTicket,
  ] as const;
  const event = events[Math.floor(Math.random() * events.length)];
  const id = Math.random().toString(36).substring(7);
  return { event, id };
}

async function publishMessages(
  nc: NatsConnection,
  subject: string,
  messageCount: number,
  delayMs: number,
) {
  const js = nc.jetstream();
  const sc = StringCodec();

  for (let i = 1; i <= messageCount; i++) {
    const message = `Mock Order #${i}`;
    await js.publish(subject, sc.encode(JSON.stringify(randomMessage())));
    console.log(`Sent message: ${message}`);
    await new Promise((resolve) => setTimeout(resolve, delayMs));
  }
}

async function main() {
  // Connect to the NATS server
  const nc = await connect({ servers: 'localhost:4222' });
  console.log('Connected to NATS');

  // Create a JetStream manager to manage streams
  const jsm = await nc.jetstreamManager();

  // Ensure the stream exists
  await setupStream(jsm, STREAM_NAME, SUBJECT_NAME);

  // Publish messages at intervals
  const messageCount = 1000;
  const delayMs = 10; // 1 second delay between messages
  await publishMessages(nc, SUBJECT_NAME, messageCount, delayMs);

  console.log('Finished sending messages.');
  await nc.close();
}

main().catch((err) => {
  console.error(`Error in sender: ${err.message}`);
});
