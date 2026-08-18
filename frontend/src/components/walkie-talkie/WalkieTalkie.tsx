import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';

interface Message {
  type: string;
  senderId: string;
  username: string;
  data: string;
  timestamp: string;
}

interface WalkieTalkieProps {
  unitId: string;
  userId: string;
  username: string;
}

export default function WalkieTalkie({ unitId, userId, username }: WalkieTalkieProps) {
  const navigate = useNavigate();
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isRecording, setIsRecording] = useState(false);
  const [activeUsers, setActiveUsers] = useState<string[]>([]);
  const [isPTTActive, setIsPTTActive] = useState(false);
  
  const wsRef = useRef<WebSocket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);

  useEffect(() => {
    connectWebSocket();
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const connectWebSocket = () => {
    const token = localStorage.getItem('token');
    if (!token) {
      toast.error('Please login first');
      navigate('/login');
      return;
    }

    const wsUrl = `ws://localhost:8080/ws?unitId=${unitId}&token=${token}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      toast.success('Connected to walkie-talkie');
      // Send join message
      ws.send(JSON.stringify({
        type: 'join',
        username: username,
        userId: userId,
        unitId: unitId
      }));
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      handleMessage(data);
    };

    ws.onclose = () => {
      setIsConnected(false);
      toast.error('Disconnected from walkie-talkie');
      setTimeout(connectWebSocket, 3000);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  };

  const handleMessage = (data: any) => {
    switch (data.type) {
      case 'room-status':
        setActiveUsers(data.users || []);
        break;
      
      case 'chat':
        setMessages(prev => [...prev, {
          type: 'chat',
          senderId: data.senderId,
          username: data.username || 'Unknown',
          data: data.data,
          timestamp: data.timestamp
        }]);
        break;

      case 'location':
        // Handle location sharing
        if (data.data) {
          const location = JSON.parse(data.data);
          toast.success(`${data.username} shared location: ${location.lat}, ${location.lng}`);
        }
        break;

      case 'welcome':
        toast.success(data.message);
        break;

      default:
        console.log('Unknown message type:', data.type);
    }
  };

  const sendMessage = () => {
    if (!inputMessage.trim() || !isConnected) return;
    
    const message = {
      type: 'chat',
      data: inputMessage,
      username: username,
      userId: userId
    };

    wsRef.current?.send(JSON.stringify(message));
    setInputMessage('');
  };

  const shareLocation = () => {
    if (!navigator.geolocation) {
      toast.error('Geolocation not supported');
      return;
    }

    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const location = {
          lat: pos.coords.latitude,
          lng: pos.coords.longitude
        };
        const message = {
          type: 'location',
          data: JSON.stringify(location),
          username: username,
          userId: userId
        };
        wsRef.current?.send(JSON.stringify(message));
        toast.success('📍 Location shared');
      },
      () => {
        toast.error('Failed to get location');
      }
    );
  };

  const togglePTT = () => {
    if (!isConnected) {
      toast.error('Not connected to walkie-talkie');
      return;
    }

    setIsPTTActive(!isPTTActive);
    
    if (!isPTTActive) {
      startRecording();
    } else {
      stopRecording();
    }
  };

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      audioContextRef.current = new AudioContext();
      mediaRecorderRef.current = new MediaRecorder(stream);
      audioChunksRef.current = [];

      mediaRecorderRef.current.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunksRef.current.push(event.data);
        }
      };

      mediaRecorderRef.current.onstop = () => {
        const audioBlob = new Blob(audioChunksRef.current, { type: 'audio/webm' });
        // Send audio as base64
        const reader = new FileReader();
        reader.onloadend = () => {
          const base64Audio = (reader.result as string).split(',')[1];
          const message = {
            type: 'signal',
            data: base64Audio,
            username: username,
            userId: userId
          };
          wsRef.current?.send(JSON.stringify(message));
        };
        reader.readAsDataURL(audioBlob);
      };

      mediaRecorderRef.current.start();
      setIsRecording(true);
      toast.success('🎤 Recording...');
    } catch (error) {
      toast.error('Failed to access microphone');
    }
  };

  const stopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
      setIsPTTActive(false);
      // Stop all tracks
      mediaRecorderRef.current.stream.getTracks().forEach(track => track.stop());
      toast.success('📻 Message sent');
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-600 to-blue-700 text-white p-4">
        <div className="flex justify-between items-center">
          <div className="flex items-center gap-2">
            <span className="text-2xl">📻</span>
            <div>
              <h2 className="font-bold">Walkie-Talkie</h2>
              <p className="text-sm text-blue-200">
                {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2 text-sm">
            <span>👥 {activeUsers.length}</span>
            <span className="text-blue-200">online</span>
          </div>
        </div>
      </div>

      {/* Active Users */}
      <div className="px-4 py-2 bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
        <div className="flex flex-wrap gap-1">
          {activeUsers.map((user, index) => (
            <span key={index} className="text-xs bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 px-2 py-1 rounded-full">
              {user}
            </span>
          ))}
          {activeUsers.length === 0 && (
            <span className="text-xs text-gray-500 dark:text-gray-400">No one else is online</span>
          )}
        </div>
      </div>

      {/* Messages */}
      <div className="h-64 overflow-y-auto p-4 bg-gray-50 dark:bg-gray-900">
        {messages.length === 0 ? (
          <div className="text-center text-gray-500 dark:text-gray-400 py-8">
            <p className="text-4xl mb-2">📻</p>
            <p>No messages yet</p>
            <p className="text-sm">Press and hold the button to talk</p>
          </div>
        ) : (
          messages.map((msg, index) => (
            <div
              key={index}
              className={`mb-2 ${msg.senderId === userId ? 'text-right' : 'text-left'}`}
            >
              <div className="inline-block max-w-[80%]">
                <div className={`text-xs text-gray-500 dark:text-gray-400 ${msg.senderId === userId ? 'text-right' : 'text-left'}`}>
                  {msg.username}
                  <span className="ml-1 text-[10px]">
                    {new Date(msg.timestamp).toLocaleTimeString()}
                  </span>
                </div>
                <div className={`px-3 py-2 rounded-lg ${
                  msg.senderId === userId
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200'
                }`}>
                  {msg.type === 'location' ? (
                    <span>📍 Location shared</span>
                  ) : (
                    <span>{msg.data}</span>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Controls */}
      <div className="p-4 border-t dark:border-gray-700 bg-white dark:bg-gray-800">
        <div className="flex gap-2">
          {/* Message Input */}
          <input
            type="text"
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && sendMessage()}
            placeholder="Type a message..."
            className="flex-1 px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            disabled={!isConnected}
          />
          <button
            onClick={sendMessage}
            disabled={!isConnected || !inputMessage.trim()}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition disabled:opacity-50"
          >
            Send
          </button>
        </div>

        <div className="flex gap-2 mt-3">
          {/* PTT Button */}
          <button
            onMouseDown={togglePTT}
            onMouseUp={togglePTT}
            onTouchStart={togglePTT}
            onTouchEnd={togglePTT}
            disabled={!isConnected}
            className={`flex-1 py-3 rounded-lg font-bold transition ${
              isPTTActive
                ? 'bg-red-600 text-white animate-pulse'
                : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
            }`}
          >
            {isPTTActive ? '🔴 Release to Send' : '🎤 Hold to Talk'}
          </button>

          {/* Location Share */}
          <button
            onClick={shareLocation}
            disabled={!isConnected}
            className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition disabled:opacity-50"
          >
            📍
          </button>

          {/* Refresh Connection */}
          <button
            onClick={connectWebSocket}
            className="px-4 py-2 bg-gray-500 text-white rounded-lg hover:bg-gray-600 transition"
          >
            🔄
          </button>
        </div>

        <div className="mt-2 text-center text-xs text-gray-500 dark:text-gray-400">
          {isRecording ? '🔴 Recording...' : 'Press and hold the red button to speak'}
        </div>
      </div>
    </div>
  );
}