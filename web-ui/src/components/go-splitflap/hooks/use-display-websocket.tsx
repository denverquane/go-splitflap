import { useState, useEffect, useRef, createContext, useContext } from 'react';

/**
 * Interface for the display state received from the WebSocket
 */
export interface DisplayState {
  activeDashboard: string;
  activeRotation: string;
  currentTime: string;
  displayState?: string; // The current display state as a string of characters
  state?: string; // Backend calls it "state" but we use "displayState" for clarity
}

interface WebSocketContextType {
  isConnected: boolean;
  displayState: DisplayState | null;
  error: string | null;
  requestState: () => void;
}

// Create a context to share the WebSocket state across components
const WebSocketContext = createContext<WebSocketContextType>({
  isConnected: false,
  displayState: null,
  error: null,
  requestState: () => {}
});

const WS_URL = import.meta.env.VITE_BACKEND_WS_URL || "ws://localhost:3000/ws"

/**
 * Global WebSocket state to be shared across the application
 */
export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const [isConnected, setIsConnected] = useState(false);
  const [displayState, setDisplayState] = useState<DisplayState | null>(null);
  const [error, setError] = useState<string | null>(null);
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // Create a WebSocket connection
    const connectWebSocket = () => {
      // Don't create a new connection if one exists
      if (socketRef.current && 
          (socketRef.current.readyState === WebSocket.OPEN || 
           socketRef.current.readyState === WebSocket.CONNECTING)) {
        return;
      }

      // For development with a proxied backend, adjust URL if needed
      const wsUrl = WS_URL;
      
      console.log('Creating new WebSocket connection...');
      const socket = new WebSocket(wsUrl);
      socketRef.current = socket;

      // Connection opened
      socket.addEventListener('open', () => {
        console.log('WebSocket connection established');
        setIsConnected(true);
        setError(null);
      });

      // Connection closed
      socket.addEventListener('close', (event) => {
        console.log('WebSocket connection closed', event);
        setIsConnected(false);
        
        // Try to reconnect after a delay, unless closed intentionally
        if (!event.wasClean) {
          setTimeout(() => {
            connectWebSocket();
          }, 3000);
        }
      });

      // Connection error
      socket.addEventListener('error', (event) => {
        console.error('WebSocket error:', event);
        setError('Connection error');
      });

      // Listen for messages
      socket.addEventListener('message', (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log('Received display state:', data);
          setDisplayState(data);
        } catch (e) {
          console.error('Error parsing WebSocket message:', e);
          setError('Error parsing server message');
        }
      });
    };

    // Start the connection
    connectWebSocket();

    // Clean up on unmount of the app (rarely happens)
    return () => {
      if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
        socketRef.current.close();
      }
    };
  }, []);

  // Function to request the current state
  const requestState = () => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      // Sending any message will cause the server to respond with the current state
      socketRef.current.send(JSON.stringify({ type: 'getState' }));
    }
  };

  // Create the context value
  const contextValue: WebSocketContextType = {
    isConnected,
    displayState,
    error,
    requestState
  };

  return (
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  );
}

/**
 * Hook to connect to the display WebSocket and receive real-time updates
 * Uses the shared WebSocketContext to avoid creating multiple connections
 */
export function useDisplayWebSocket() {
  const context = useContext(WebSocketContext);
  return context;
}

/**
 * Hook to get the active state information from the WebSocket
 * @returns Object containing the active dashboard and rotation names
 */
export function useActiveState() {
  const { displayState, isConnected } = useDisplayWebSocket();
  
  return {
    activeDashboard: displayState?.activeDashboard || "",
    activeRotation: displayState?.activeRotation || "",
    isConnected
  };
}