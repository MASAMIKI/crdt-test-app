'use client'

import * as Y from 'yjs'
import { WebsocketProvider } from 'y-websocket'

import {useEffect, useState} from "react";

import 'react-toastify/dist/ReactToastify.css';

const Karte = () => {
  const projectId = "01H7ZD6V8S7XE3A2HHRD6J0TSG" // 仮のプロジェクトID
  const [wsProvider, setWsProvider] = useState<WebsocketProvider>(null);
  const [doc, setDoc] = useState<Y.Doc>(null);

  const [isConnected, setIsConnected] = useState(false);
  const [input, setInput] = useState('');
  const [userId, setUserId] = useState('');

  const [yText, setYText] = useState<Y.Text>(null);

  useEffect(() => {
    const doc = new Y.Doc()
    setDoc(doc);

    return () => {
      if (wsProvider) wsProvider.destroy();
    };
  }, []);


  const joinRoom = () => {
    // WebSocketの接続を確立
    const wsProvider = new WebsocketProvider(
`ws://localhost:8080/project/karte`,
      projectId,
      doc
    )
    wsProvider.on('status', event => {
      if (event.status === 'connected') {
        setIsConnected(true);
      } else {
        setIsConnected(false);
      }
    })
    setWsProvider(wsProvider);

    // Yjsのテキストタイプを作成
    const yText = doc.getText('message');
    yText.observe(event => {
      // yTextが変更されるたびに、このコードが実行されます
      const updatedContent = yText.toString();
      setInput(updatedContent)
    });

    setYText(yText)
  }

  const leaveRoom = () => {
    if (wsProvider) wsProvider.destroy();
    setWsProvider(null);
    setInput('');
  }

  const handleInputChange = (e) => {
    const newValue = e.target.value;
    setInput(newValue);

    yText.delete(0, yText.toString().length);
    yText.insert(0, newValue);
  };

  return (
    <>
      <p>状態: {isConnected ? "接続中" : "切断中"}</p>
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
            <button className={`mt-7 text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm w-full sm:w-auto px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800 disabled:opacity-50`}
                    onClick={isConnected ? leaveRoom : joinRoom }
                    type="button"
                    disabled={!userId}
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
                 onChange={handleInputChange}
          />
        </div>
      </div>
    </>
  )
}

export default Karte