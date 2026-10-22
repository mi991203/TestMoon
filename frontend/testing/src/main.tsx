import React from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { Providers } from "@testmoon/web";
import { App } from "./app/App";
createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <Providers>
      <BrowserRouter basename="/apps/testing">
        <App />
      </BrowserRouter>
    </Providers>
  </React.StrictMode>,
);
