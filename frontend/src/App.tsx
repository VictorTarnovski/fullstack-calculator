import { Toaster } from '@/components/ui/sonner'
import IndexPage from '@/pages/IndexPage'

export default function App() {
  return (
    <>
      <IndexPage />
      {/* Mounted once at the root so any component can raise a toast. */}
      <Toaster position="top-center" richColors />
    </>
  )
}
