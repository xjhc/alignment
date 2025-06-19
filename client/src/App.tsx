import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)

  return (
    <div className="App">
      <header className="App-header">
        <h1>Alignment</h1>
        <p>Corporate social deduction game</p>
        <div className="card">
          <button onClick={() => setCount((count) => count + 1)}>
            count is {count}
          </button>
          <p>
            Game implementation coming soon...
          </p>
        </div>
      </header>
    </div>
  )
}

export default App