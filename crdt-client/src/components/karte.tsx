'use client'

import {useEffect, useState} from "react";
import {ToastContainer, toast} from "react-toastify";

import 'react-toastify/dist/ReactToastify.css';

const Karte = () => {
  const [ws, setWs] = useState<WebSocket>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');

  useEffect(() => {
    // WebSocketの接続を確立
    const websocket = new WebSocket('ws://localhost:8080/ws');

    websocket.onopen = () => {
      setIsConnected(true);
      toast(`WebSocket connected!`)
    };

    websocket.onclose = () => {
      setIsConnected(false);
      toast(`WebSocket disconnected!`)
    }

    websocket.onmessage = (event) => {
      setMessages(prevMessages => [...prevMessages, event.data]);
    };

    setWs(websocket);

    // コンポーネントのアンマウント時にWebSocketをクローズ
    return () => {
      websocket.close();
    };
  }, []);

  const handleSend = () => {
    if (ws) {
      ws.send(input);
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
            <label htmlFor="message1" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
              Message 1
            </label>
            <input type="text" id="message1"
                   className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                   value={input}
                   onChange={(e) => setInput(e.target.value)}
            />
          </div>
          <div>
            <label htmlFor="message2" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
              Message 2
            </label>
            <input type="text" id="message2"
                   className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
            />
          </div>
        </div>
        <div className="mb-6">
          <label htmlFor="message3" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
            Message 3
          </label>
          <input type="text" id="message3"
                 className="bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
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