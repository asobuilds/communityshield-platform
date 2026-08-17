import { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';

interface WalkieTalkieProps {
  unitId: string;
  userId: string;
  userName: string;
}

function WalkieTalkie({ unitId, userId, userName }: WalkieTalkieProps) {
  const [messages, setMessages] = useState<any[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(true);
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [isPushing, setIsPushing] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const room = `unit:${unitId}`;

  useEffect(() => {
    fetchHistory();
    const wsUrl = `ws://localhost:8080/api/v1/ws?userId=${userId}&room=${room}`;
    const socket = new WebSocket(wsUrl);
    setWs(socket);

    socket.onopen = () => console.log('Walkie WS connected');
    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      setMessages(prev => [...prev, msg]);
    };
    socket.onclose = () => console.log('Walkie WS disconnected');

    return () => socket.close();
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const fetchHistory = async () => {
    try {
      const response = await axios.get(`/api/v1/chat/history?room=${room}`);
      setMessages(response.data.messages || []);
    } catch (error) {
      console.error('Failed to fetch walkie history');
    } finally {
      setLoading(false);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const sendMessage = (content: string) => {
    if (!content.trim() || !ws) return;
    const msg = {
      senderId: userId,
      senderName: userName,
      content: content,
      room: room,
      type: 'walkie',
      timestamp: new Date().toISOString(),
    };
    ws.send(JSON.stringify(msg));
  };

  const handlePushStart = () => {
    setIsPushing(true);
    toast.info('Push-to-talk: Press to speak (text mode for demo)');
  };

  const handlePushEnd = () => {
    setIsPushing(false);
    if (input.trim()) {
      sendMessage(input);
      setInput('');
    }
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow border dark:border-gray-700 p-4">
      <h2 className="text-xl font-bold text-blue-600 dark:text-blue-400 mb-4">📻 Walkie-Talkie</h2>

      <div className="flex-1 overflow-y-auto h-64 p-2 space-y-2 border rounded-lg dark:border-gray-600 bg-gray-50 dark:bg-gray-900">
        {loading ? (
          <div className="text-center text-gray-500">Loading...</div>
        ) : messages.length === 0 ? (
          <div className="text-center text-gray-500">No messages yet. Press the button to talk!</div>
        ) : (
          messages.map((msg, idx) => (
            <div
              key={idx}
              className={`flex ${msg.senderId === userId ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-xs px-4 py-2 rounded-lg ${
                  msg.senderId === userId
                    ? 'bg-green-600 text-white'
                    : 'bg-yellow-200 dark:bg-yellow-800 text-gray-800 dark:text-gray-200'
                }`}
              >
                <p className="text-sm font-semibold">{msg.senderName || 'Officer'}</p>
                <p className="text-sm">{msg.content}</p>
                <p className="text-xs opacity-70 mt-1">{formatTime(msg.timestamp)}</p>
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      <div className="mt-4 flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Type message or press button..."
          className="flex-1 px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
        />
        <button
          onClick={() => {
            if (input.trim()) {
              sendMessage(input);
              setInput('');
            }
          }}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
        >
          Send
        </button>
        <button
          onMouseDown={handlePushStart}
          onMouseUp={handlePushEnd}
          onTouchStart={handlePushStart}
          onTouchEnd={handlePushEnd}
          className={`px-6 py-2 rounded-lg font-bold transition ${
            isPushing
              ? 'bg-red-600 text-white scale-95'
              : 'bg-gray-600 text-white hover:bg-gray-700'
          }`}
        >
          {isPushing ? '🔴 PUSHING' : '🎙️ PTT'}
        </button>
      </div>
      <p className="text-xs text-gray-500 mt-2">Press and hold PTT to send (text mode for demo).</p>
    </div>
  );
}

export default WalkieTalkie;