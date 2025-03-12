import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { Routes } from "@generouted/react-router";

import { Providers } from "./providers";

import "./main.css";

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <Providers>
            <Routes/>
        </Providers>
    </StrictMode>,
);
