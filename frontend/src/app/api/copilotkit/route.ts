import {
  CopilotRuntime,
  OpenAIAdapter,
  copilotRuntimeNextJSAppRouterEndpoint,
} from "@copilotkit/runtime";
import { NextRequest } from "next/server";

// Create a simple adapter that works with our AG-UI service
class HealthSecureAdapter extends OpenAIAdapter {
  constructor() {
    super({
      model: "gpt-4",
      apiKey: "dummy-key", // We don't use this since we override process
    });
  }

  async process(forwardedProps: any) {
    try {
      // Extract messages from the forwarded props
      const messages = forwardedProps.messages || [];
      const lastMessage = messages[messages.length - 1];

      if (!lastMessage || !lastMessage.content) {
        return {
          content: "Hello! How can I help you with medical information today?",
          role: "assistant",
        };
      }

      // Call our AG-UI service
      const response = await fetch("http://localhost:5000/agui/chat", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          message: lastMessage.content,
          thread_id: "copilotkit_" + Date.now(),
          user_id: "frontend_user",
          user_role: "admin",
          user_name: "Frontend User",
          model: "gpt-oss:20b-cloud",
          history: messages.slice(0, -1)
        }),
      });

      if (response.ok) {
        const data = await response.json();
        return {
          content: data.response || "I'm ready to help with medical questions.",
          role: "assistant",
        };
      } else {
        throw new Error("AI service unavailable");
      }

    } catch (error) {
      console.error("HealthSecure AI error:", error);
      return {
        content: "I'm currently unable to connect to the AI service. Please try again in a moment.",
        role: "assistant",
      };
    }
  }
}

const serviceAdapter = new HealthSecureAdapter();
const runtime = new CopilotRuntime();

export const POST = async (req: NextRequest) => {
  const { handleRequest } = copilotRuntimeNextJSAppRouterEndpoint({
    runtime: runtime,
    serviceAdapter: serviceAdapter,
    endpoint: "/api/copilotkit",
  });

  return handleRequest(req);
};