import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { applyMode, Mode } from "@cloudscape-design/global-styles";

import App from "./App.tsx";

const darkQuery = window.matchMedia("(prefers-color-scheme: dark)");
applyMode(darkQuery.matches ? Mode.Dark : Mode.Light);
darkQuery.addEventListener("change", (e) => {
  applyMode(e.matches ? Mode.Dark : Mode.Light);
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
