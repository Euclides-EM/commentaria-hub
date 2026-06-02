export const colorFromText = (input: string) => {
  let hash = 0
  const multiplier = 53 + 17

  for (let i = 0; i < input.length; i++) {
    hash = input.charCodeAt(i) + hash * multiplier
    hash = (hash << 5) + hash + input.charCodeAt(i)
    hash = (hash << 5) + hash + input.charCodeAt(i)
    hash = (hash << 5) + hash + input.charCodeAt(i)
  }

  const h = Math.abs(hash + 200) % 360

  const s = 60 + (Math.abs(hash >> 8) % 20)

  const l = 55 + (Math.abs(hash >> 16) % 15)

  return `hsl(${h}, ${s}%, ${l}%)`
}
