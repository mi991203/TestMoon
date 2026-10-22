import React from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Providers } from "@testmoon/web";
import { Portal } from "./app/Portal";

import { SharedDocument } from "./features/SharedDocument";
createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <Providers>
      <BrowserRouter>
        <Routes>
          <Route path="/share/:token" element={<SharedDocument />} />
          <Route path="*" element={<Portal />} />
        </Routes>
      </BrowserRouter>
    </Providers>
  </React.StrictMode>,
);
