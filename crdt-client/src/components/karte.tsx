'use client'

import {useEffect, useState} from "react";
import {ToastContainer, toast} from "react-toastify";

import 'react-toastify/dist/ReactToastify.css';

type SendMessage = {
  sentBy: string;
  content: string;
}

type ReceiveMessage = {
  sentBy: string;
  content: string;
}

const Karte = () => {
  const projectId = "01H7ZD6V8S7XE3A2HHRD6J0TSG" // 仮のプロジェクトID
  const [ws, setWs] = useState<WebSocket>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [userId, setUserId] = useState('');

  useEffect(() => {
    return () => {
      if (ws) ws.close();
    };
  }, []);

  const onOpen = () => {
    setIsConnected(true);
    toast(`WebSocket connected!`)
  }

  const onClose = () => {
    toast(isConnected ? `WebSocket disconnected!` : 'WebSocket connection failed!')
    setIsConnected(false);
  }

  const onError = (event: Event) => {
    console.error(event);
  }

  const onMessage = (event: MessageEvent) => {
    event.data.text().then((text) => {
      const data: ReceiveMessage = JSON.parse(text);
      setMessages(prevMessages => [...prevMessages, `${data.sentBy}:${data.content}`]);
    }).catch((error) => {
      console.error(error);
    })
  }

  const joinRoom = () => {
    // WebSocketの接続を確立
    const websocket = new WebSocket(`ws://localhost:8080/project/${projectId}/karte?userId=${userId}`);
    websocket.onopen = onOpen;
    websocket.onclose = onClose
    websocket.onerror = onError
    websocket.onmessage = onMessage
    setWs(websocket);
  }

  const leaveRoom = () => {
    ws.close();
    setWs(null);
    setMessages([])
  }

  const handleSend = () => {
    if (ws) {
      const data: SendMessage = {
        content: input,
        sentBy: userId,
      }
      const text = JSON.stringify(data);
      const binaryData = new TextEncoder().encode(text);
      ws.send(binaryData);
      setInput(''); // 入力フィールドをクリア
    }
  };

  return (
    <>
      <ToastContainer
        position="top-right"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
        theme="light"
      />
      <p>状態: {isConnected ? "接続中" : "切断中"}</p>
      <div className="w-1/3">
        <label htmlFor="messages" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
          Messages (readOnly)
        </label>
        <textarea id="messages" rows="4"
                  className="block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                  value={messages.join('\n')}
                  readOnly
        >
        </textarea>
      </div>
      <div className="w-1/3">
        <div className="grid gap-6 mb-6 md:grid-cols-2">
          <div>
            <label htmlFor="userId" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
              ユーザーID
            </label>
            <input type="text" id="userId"
                   className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                   value={userId}
                   onChange={(e) => setUserId(e.target.value)}
            />
          </div>
          <div>
            <button className="mt-7 text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                    onClick={isConnected ? leaveRoom : joinRoom }
                    type="button"
            >
              { isConnected ? '退室' : '入室'}
            </button>
          </div>
        </div>
        <div className="mb-6">
          <label htmlFor="message" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
            Message
          </label>
          <input type="text" id="message"
                 className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                 value={input}
                 onChange={(e) => setInput(e.target.value)}
          />
        </div>
        <button className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
                onClick={handleSend}
                type="button"
        >
          Submit
        </button>
      </div>
    </>
  )
}

export default Karte