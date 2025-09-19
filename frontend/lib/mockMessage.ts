export const askMahjongAI = async (prompt: string) => {
  const response = await fetch(
    "http://localhost:8081/mahjong.ai.v1.MahjongAIService/AskMahjongAI",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ prompt, max_tokens: 2000, temperature: 0.7 }),
    }
  );
  return response.json();
};
