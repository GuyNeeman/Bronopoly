import { useState } from 'react'
import './App.css'
import StartScreen from './components/StartScreen'
import PlayerScreen from './components/PlayerScreen.jsx'
import SettingsScreen from './components/SettingsScreen'
import OnlineScreen from "./components/OnlineScreen.jsx";
import PlayerCreationScreen from "./components/PlayerCreationScreen.jsx";

function App() {
    const [screen, setScreen] = useState("start");

    const [playerCount, setPlayerCount] = useState(2);

    const [players, setPlayers] = useState([
    ]);

    const [settings, setSettings] = useState({
        startMoney: 2000,
        roundMoney: 200,
        vacationCash: true,
        auction: true,
        noRentInPrison: true,
        evenBuild: true,
    });

    function startGame() {
        console.log("Start game:", { playerCount, settings });
    }

    return (
        <div className="app">
            {screen === "start" && (
                <StartScreen setScreen={setScreen} />
            )}

            {screen === "players" && (
                <PlayerScreen
                    playerCount={playerCount}
                    setPlayerCount={setPlayerCount}
                    next={() => setScreen("settings")}
                    back={() => setScreen("start")}
                />
            )}

            {screen === "settings" && (
                <SettingsScreen
                    settings={settings}
                    setSettings={setSettings}
                    startGame={startGame}
                    next={() => setScreen("create-players")}
                    back={() => setScreen("players")}
                />
            )}

            {screen === "online" && (
                <OnlineScreen
                    back={() => setScreen("start")}
                />
            )}

            {screen === "create-players" && (
                <PlayerCreationScreen
                    playerCount={playerCount}
                    players={players}
                    setPlayers={setPlayers}
                    next={() => setScreen("settings")}
                    back={() => setScreen("players")}
                    startMoney={settings.startMoney}
                />
            )}
        </div>
    );
}

export default App;