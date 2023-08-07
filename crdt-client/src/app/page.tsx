import Karte from "@/components/karte";
import { ToastContainer } from 'react-toastify';

export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-between p-24">
      建物カルテ
      <Karte />
    </main>
  )
}
