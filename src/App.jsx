import { useState } from 'react'
import './App.css'
import {Slider} from "@mui/material";

function App() {
    const [playercount, setPlayerCount] = useState("123");
    const [welcome, setWelcome] = useState(false);
    const [online, setOnline] = useState(false);

    function createGame() {
        console.log("hi")
    }
    function createPlayer(playercount) {
        console.log("hi")
        }

  return (
    <>
        {!welcome && (
            <>
                <p>Welcome to Bronopoly</p>
                <p>Will you play online(currently unavailable) or offline </p>
                <div>
                    <button onClick={() => {setOnline(true);setWelcome(true);}}>
                        Online
                    </button>

                    <button onClick={() => {setOnline(false);setWelcome(true);}}>
                        Offline
                    </button>
                </div>
            </>
        )}
        {welcome && online && (
            <>
                <p>No Online Mode currently available.</p>
            </>
        )}
        {welcome && !online && (
            <>
                <p>Great to know!</p>
                <p>How many players will be joining you?</p>
                <Slider
                    size="small"
                    defaultValue={0}
                    aria-label="Small"
                    valueLabelDisplay="auto"
                    max={5}
                />
            </>
        )}
    </>
  )
}

export default App
