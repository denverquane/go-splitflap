import { ReactNode } from "react";



import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { initTsrReactQuery } from "@ts-rest/react-query/v5";



import { ThemeProvider } from "./components/shadcn/hooks/theme-provider";

import { WebSocketProvider } from "./components/go-splitflap/hooks";

import { displayContract } from "./lib/contract";

const API_URL = import.meta.env.VITE_BACKEND_API_URL || "http://localhost:3000"


const queryClient = new QueryClient();

export const client = initTsrReactQuery(displayContract, {

    baseUrl: API_URL,

    baseHeaders: {},

});



export function Providers({ children }: { children: ReactNode }) {

    return (

        <QueryClientProvider client={queryClient}>

            <client.ReactQueryProvider>

                <ThemeProvider defaultTheme="system" storageKey="skysync-theme">

                    <WebSocketProvider>

                        {children}

                    </WebSocketProvider>

                </ThemeProvider>

            </client.ReactQueryProvider>

        </QueryClientProvider>

    );

}