import { useState } from 'react'
import './App.css'
import StartScreen from './screens/StartScreen.jsx'
import PlayerScreen from './components/PlayerScreen.jsx'
import SettingsScreen from './screens/SettingsScreen.jsx'
import JoinScreen from "./screens/JoinScreen.jsx";
import PlayerCreationScreen from "./components/PlayerCreationScreen.jsx";
import StatsticScreen from "./screens/StatisticScreen.jsx";

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

            {screen === "join" && (
                <JoinScreen
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

            {screen === "statistic" && (
                <StatsticScreen
                    back={() => setScreen("start")}
                />
            )}
        </div>
    );
}

export default App;