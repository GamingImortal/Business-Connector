export default function Avatar({ emoji, color, size = 49 }) {
  return (
    <div className="avatar" style={{ background: color, width: size, height: size, fontSize: size * 0.5 }}>
      {emoji}
    </div>
  )
}
