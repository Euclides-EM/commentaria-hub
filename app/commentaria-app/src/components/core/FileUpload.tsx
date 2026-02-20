import { useRef } from 'react'
import { Button } from './Button'

type FileUploadProps = {
  file: File | null
  onFileChange: (file: File | null) => void
  accept?: string
  disabled?: boolean
  required?: boolean
  buttonLabel?: string
  className?: string
}

export function FileUpload({
  file,
  onFileChange,
  accept,
  disabled = false,
  required = false,
  buttonLabel = 'Choose file',
  className,
}: FileUploadProps) {
  const inputRef = useRef<HTMLInputElement | null>(null)

  return (
    <div className={`flex gap-2 items-center w-fit ${className || ''}`}>
      <Button
        className="p-2"
        onClick={(event) => {
          event.preventDefault()
          event.stopPropagation()
          inputRef.current?.click()
        }}
        disabled={disabled}
      >
        {buttonLabel}
      </Button>
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        onChange={(event) => onFileChange(event.target.files?.[0] || null)}
        className="sr-only"
        disabled={disabled}
        required={required}
      />
      {file && (
        <p className="text-sm text-gray-500">
          Selected:{' '}
          <span className="font-bold text-black italic">{file.name}</span>
        </p>
      )}
    </div>
  )
}
