export const parseRichText = (text: string): string => {
  if (!text) return "";

  let escapedText = text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");

  escapedText = escapedText.replace(
    /#([a-zA-Z0-9_]+)/g,
    '<span class="hashtag-link" data-tag="$1" style="color: #00376b; cursor: pointer; font-weight: 400;">#$1</span>'
  );

  escapedText = escapedText.replace(
    /@([a-zA-Z0-9._]+)/g,
    '<span class="mention-link" data-username="$1" style="color: #0095f6; cursor: pointer; font-weight: 600;">@$1</span>'
  );

  return escapedText;
};
