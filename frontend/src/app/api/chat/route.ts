import { NextRequest } from "next/server";

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const { messages } = body;

    if (!messages || !Array.isArray(messages)) {
      return Response.json({
        error: "Messages array is required"
      }, { status: 400 });
    }

    const lastMessage = messages[messages.length - 1];

    if (!lastMessage || !lastMessage.content) {
      return Response.json({
        response: "Hello! How can I help you with medical information today?"
      });
    }

    // Call our AG-UI service directly
    const response = await fetch("http://localhost:5000/agui/chat", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        message: lastMessage.content,
        thread_id: "chat_" + Date.now(),
        user_id: "frontend_user",
        user_role: "admin",
        user_name: "Frontend User",
        model: "gpt-oss:20b-cloud",
        history: messages.slice(0, -1)
      }),
    });

    if (response.ok) {
      const data = await response.json();
      return Response.json({
        response: data.response || "I'm ready to help with medical questions."
      });
    } else {
      throw new Error("AI service unavailable");
    }

  } catch (error) {
    console.error("Chat API error:", error);
    return Response.json({
      response: "I'm currently unable to connect to the AI service. Please try again in a moment."
    });
  }
}