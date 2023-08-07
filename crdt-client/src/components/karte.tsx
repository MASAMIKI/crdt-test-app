'use client'

import {useEffect, useState} from "react";
import {ToastContainer, toast} from "react-toastify";
import {socket} from "@/utils/socket";

import 'react-toastify/dist/ReactToastify.css';

const Karte = () => {
  const [isConnected, setIsConnected] = useState(socket.connected);
  const [fooEvents, setFooEvents] = useState([]);
  const [message, setMessage] = useState("");
  const [messages, setMessages] = useState([]);

  useEffect(() => {
    // 各種イベントの定義
    const onConnect = () => {
      setIsConnected(true);
      toast(`${socket.id}: Connected to WebSocket`)
    }
    const onDisconnect = () => {
      setIsConnected(false);
      toast(`${socket.id}: Disconnected from WebSocket`)
    }

    const onFooEvent = (value) => {
      setFooEvents(previous => [...previous, value]);
    }

    // 接続
    socket.connect()
    console.log(socket);
    socket.on("connect", onConnect);
    socket.on("disconnect", onDisconnect);
    socket.on('foo', onFooEvent);
    socket.on("connect_error", (e) => {
      console.log(e);
    });

    // クリーンアップ
    return () => {
      socket.disconnect()
      socket.off('connect', onConnect);
      socket.off('disconnect', onDisconnect);
      socket.off('foo', onFooEvent);
    };
  }, []);

  const handleSend = () => {
    console.log(message);
    socket.timeout(5000).emit("message", message, () => {
      toast(`${socket.id}: Sent message: ${message}`)
    });
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
        {messages.map((msg, index) => (
          <div key={index}>{msg}</div>
        ))}

        <label htmlFor="messages" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">
          Messages (readOnly)
        </label>
        <textarea id="messages" rows="4"
                  className="block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                  value={messages.join("\n")}
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
                   value={message}
                   onChange={(e) => setMessage(e.target.value)}
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