export const generateMahjongResponse = (): string => {
  const responses = [
    "麻雀の基本戦術として、まず安全牌を意識することが重要です。相手の捨て牌をよく観察しましょう。",
    "リーチをかけるタイミングは、手牌の安全性と点数効率を考慮して決めましょう。",
    "役作りでは、タンヤオやピンフなどの基本役から覚えることをお勧めします。",
    "守備では、現物や筋を意識して安全牌を選ぶことが大切です。",
    "点数計算は慣れが必要ですが、基本的な符計算から始めてみてください。",
  ];
  return responses[Math.floor(Math.random() * responses.length)];
};
