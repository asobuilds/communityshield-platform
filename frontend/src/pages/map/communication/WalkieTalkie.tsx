import { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';

interface Message {
  id: string;
  message: string;
  messageType: string;
  senderId: string;
  sender: { firstName: string; lastName: string };
  priority: string;
  isEmergency: boolean;
  createdAt: string;
  reactions?: string;
  fileUrl?: string;
  fileName?: string;
}

interface WalkieTalkieProps {
  unitId: string;
  userId: string;
  username: string;
  roomId?: string;
}

export default function WalkieTalkie({ unitId, userId, username, roomId }: WalkieTalkieProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [priority, setPriority] = useState<'normal' | 'high' | 'emergency'>('normal');
  const [isEmergency, setIsEmergency] = useState(false);
  const [activeUsers, setActiveUsers] = useState<string[]>([]);
  const [rooms, setRooms] = useState<any[]>([]);
  const [selectedRoom, setSelectedRoom] = useState(roomId || '');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchRooms();
    if (selectedRoom) {
      fetchMessages();
    }
  }, [selectedRoom]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const fetchRooms = async () => {
    try {
      const response = await axios.get(`/api/v1/communication/units/${unitId}/rooms`);
      setRooms(response.data.rooms || []);
      if (!selectedRoom && response.data.rooms.length > 0) {
        setSelectedRoom(response.data.rooms[0].id);
      }
    } catch (error) {
      console.error('Failed to fetch rooms');
    }
  };

  const fetchMessages = async () => {
    if (!selectedRoom) return;
    setLoading(true);
    try {
      const response = await axios.get(`/api/v1/communication/rooms/${selectedRoom}/messages`);
      setMessages(response.data.messages || []);
    } catch (error) {
      toast.error('Failed to fetch messages');
    } finally {
      setLoading(false);
    }
  };

  const sendMessage = async () => {
    if (!inputMessage.trim() || !selectedRoom) return;

    try {
      const response = await axios.post('/api/v1/communication/messages', {
        roomId: selectedRoom,
        message: inputMessage,
        messageType: 'text',
        priority: priority,
        isEmergency: isEmergency,
      });

      setMessages([...messages, response.data.data]);
      setInputMessage('');
      setIsEmergency(false);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to send message');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'emergency': return 'bg-red-600 text-white';
      case 'high': return 'bg-orange-600 text-white';
      default: return 'bg-blue-600 text-white';
    }
  };

  const getPriorityEmoji = (priority: string) => {
    switch (priority) {
      case 'emergency': return '🚨';
      case 'high': return '🔴';
      default: return '📩';
    }
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

      {/* Room Selector */}
      <div className="px-4 py-2 bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
        <select
          value={selectedRoom}
          onChange={(e) => setSelectedRoom(e.target.value)}
          className="w-full px-3 py-1 border rounded-lg dark:bg-gray-600 dark:border-gray-500 dark:text-white text-sm"
        >
          {rooms.map((room) => (
            <option key={room.id} value={room.id}>{room.name}</option>
          ))}
        </select>
      </div>

      {/* Messages */}
      <div className="h-80 overflow-y-auto p-4 bg-gray-50 dark:bg-gray-900">
        {loading ? (
          <div className="text-center py-8 text-gray-500">Loading messages...</div>
        ) : messages.length === 0 ? (
          <div className="text-center text-gray-500 dark:text-gray-400 py-8">
            <p className="text-4xl mb-2">📻</p>
            <p>No messages yet</p>
            <p className="text-sm">Start the conversation</p>
          </div>
        ) : (
          messages.map((msg) => (
            <div
              key={msg.id}
              className={`mb-3 ${msg.senderId === userId ? 'text-right' : 'text-left'}`}
            >
              <div className="inline-block max-w-[80%]">
                <div className={`text-xs text-gray-500 dark:text-gray-400 ${msg.senderId === userId ? 'text-right' : 'text-left'}`}>
                  <span className="font-medium">{msg.sender?.firstName} {msg.sender?.lastName}</span>
                  <span className="ml-1 text-[10px]">
                    {new Date(msg.createdAt).toLocaleTimeString()}
                  </span>
                  {msg.priority && msg.priority !== 'normal' && (
                    <span className={`ml-1 px-1 py-0.5 rounded text-[10px] ${getPriorityColor(msg.priority)}`}>
                      {getPriorityEmoji(msg.priority)} {msg.priority}
                    </span>
                  )}
                </div>
                <div className={`px-3 py-2 rounded-lg text-sm ${
                  msg.senderId === userId
                    ? 'bg-blue-600 text-white'
                    : msg.isEmergency
                    ? 'bg-red-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-200'
                }`}>
                  {msg.message}
                  {msg.fileUrl && (
                    <div className="mt-1">
                      <a href={msg.fileUrl} target="_blank" rel="noopener noreferrer" className="text-xs underline">
                        📎 {msg.fileName || 'Attachment'}
                      </a>
                    </div>
                  )}
                </div>
                {msg.reactions && (
                  <div className="text-xs mt-1 text-gray-400">
                    {msg.reactions}
                  </div>
                )}
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Controls */}
      <div className="p-4 border-t dark:border-gray-700 bg-white dark:bg-gray-800">
        <div className="flex gap-2">
          {/* Priority Selector */}
          <select
            value={priority}
            onChange={(e) => setPriority(e.target.value as any)}
            className="px-2 py-2 border rounded-lg text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          >
            <option value="normal">Normal</option>
            <option value="high">High</option>
            <option value="emergency">🚨 Emergency</option>
          </select>

          {/* Message Input */}
          <input
            type="text"
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder="Type a message..."
            className="flex-1 px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
          />

          <button
            onClick={sendMessage}
            disabled={!inputMessage.trim()}
            className={`px-4 py-2 rounded-lg text-white transition ${
              isEmergency
                ? 'bg-red-600 hover:bg-red-700'
                : priority === 'high'
                ? 'bg-orange-600 hover:bg-orange-700'
                : 'bg-blue-600 hover:bg-blue-700'
            } disabled:opacity-50`}
          >
            Send
          </button>
        </div>

        <div className="flex gap-2 mt-2">
          <button
            onClick={() => setIsEmergency(!isEmergency)}
            className={`px-3 py-1 rounded text-sm transition ${
              isEmergency
                ? 'bg-red-600 text-white animate-pulse'
                : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
            }`}
          >
            🚨 {isEmergency ? 'Emergency On' : 'Emergency Off'}
          </button>

          <button
            onClick={fetchMessages}
            className="px-3 py-1 rounded text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 transition"
          >
            🔄 Refresh
          </button>
        </div>
      </div>
    </div>
  );
}