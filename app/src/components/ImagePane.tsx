import { useRef } from 'react'

interface ImagePaneProps {
  imageUrl: string
  imgStatus: string
}

export function ImagePane({ imageUrl, imgStatus }: ImagePaneProps) {
  const pageImgRef = useRef<HTMLImageElement>(null)

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Image</div>
        <div className="text-xs opacity-75">{imgStatus}</div>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden flex items-center justify-center p-0">
        <img
          ref={pageImgRef}
          id="pageImg"
          alt="Page image"
          className="max-w-full max-h-full w-auto h-auto object-contain block"
          src={imageUrl}
        />
      </div>
    </section>
  )
}
