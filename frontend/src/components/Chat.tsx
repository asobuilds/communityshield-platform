import { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';

interface ChatProps {
  room: string;
  userId: string;
  userName: string;
}

function Chat({ room, userId, userName }: ChatProps) {
  const [messages, setMessages] = useState<any[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(true);
  const [ws, setWs] = useState<WebSocket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchHistory();
    const wsUrl = `ws://localhost:8080/api/v1/ws?userId=${userId}&room=${room}`;
    const socket = new WebSocket(wsUrl);
    setWs(socket);

    socket.onopen = () => console.log('WebSocket connected');
    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      setMessages(prev => [...prev, msg]);
    };
    socket.onclose = () => console.log('WebSocket disconnected');

    return () => socket.close();
  }, [room]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const fetchHistory = async () => {
    try {
      const response = await axios.get(`/api/v1/chat/history?room=${room}`);
      setMessages(response.data.messages || []);
    } catch (error) {
      console.error('Failed to fetch history');
    } finally {
      setLoading(false);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const sendMessage = () => {
    if (!input.trim() || !ws) return;
    const msg = {
      senderId: userId,
      senderName: userName,
      content: input,
      room: room,
      type: 'chat',
      timestamp: new Date().toISOString(),
    };
    ws.send(JSON.stringify(msg));
    setInput('');
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="flex flex-col h-96 bg-white dark:bg-gray-800 rounded-lg shadow border dark:border-gray-700">
      <div className="flex-1 overflow-y-auto p-4 space-y-2">
        {loading ? (
          <div className="text-center text-gray-500">Loading...</div>
        ) : messages.length === 0 ? (
          <div className="text-center text-gray-500">No messages yet. Start chatting!</div>
        ) : (
          messages.map((msg, idx) => (
            <div
              key={idx}
              className={`flex ${msg.senderId === userId ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-xs px-4 py-2 rounded-lg ${
                  msg.senderId === userId
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200'
                }`}
              >
                <p className="text-sm font-semibold">{msg.senderName || 'User'}</p>
                <p className="text-sm">{msg.content}</p>
                <p className="text-xs opacity-70 mt-1">{formatTime(msg.timestamp)}</p>
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>
      <div className="p-3 border-t dark:border-gray-700 flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
          placeholder="Type a message..."
          className="flex-1 px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
        />
        <button
          onClick={sendMessage}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
        >
          Send
        </button>
      </div>
    </div>
  );
}

export default Chat;