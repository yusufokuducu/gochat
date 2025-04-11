class WebSocketService {
  constructor() {
    this.socket = null;
    this.messageHandlers = [];
    this.connectionHandlers = [];
    this.errorHandlers = [];
  }

  connect(token) {
    if (this.socket) {
      this.disconnect();
    }

    const wsUrl = `${process.env.REACT_APP_API_URL?.replace('http', 'ws') || 'ws://localhost:8000/api'}/ws/chat/${token}`;
    this.socket = new WebSocket(wsUrl);

    this.socket.onopen = () => {
      console.log('WebSocket connection established');
      this.connectionHandlers.forEach(handler => handler(true));
    };

    this.socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.messageHandlers.forEach(handler => handler(data));
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.errorHandlers.forEach(handler => handler(error));
    };

    this.socket.onclose = () => {
      console.log('WebSocket connection closed');
      this.connectionHandlers.forEach(handler => handler(false));
    };
  }

  disconnect() {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  sendMessage(message) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message));
    } else {
      console.error('WebSocket is not connected');
    }
  }

  onMessage(handler) {
    this.messageHandlers.push(handler);
    return () => {
      this.messageHandlers = this.messageHandlers.filter(h => h !== handler);
    };
  }

  onConnectionChange(handler) {
    this.connectionHandlers.push(handler);
    return () => {
      this.connectionHandlers = this.connectionHandlers.filter(h => h !== handler);
    };
  }

  onError(handler) {
    this.errorHandlers.push(handler);
    return () => {
      this.errorHandlers = this.errorHandlers.filter(h => h !== handler);
    };
  }

  isConnected() {
    return this.socket && this.socket.readyState === WebSocket.OPEN;
  }
}

// Create a singleton instance
const websocketService = new WebSocketService();

export default websocketService;
