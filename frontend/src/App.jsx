import { useState, useEffect } from 'react';
import { Routes, Route, useNavigate } from 'react-router-dom';
import './App.css';
import StartScreen from './screens/StartScreen.jsx';
import PlayerScreen from './components/PlayerScreen.jsx';
import SettingsScreen from './screens/SettingsScreen.jsx';
import JoinScreen from "./screens/JoinScreen.jsx";
import PlayerCreationScreen from "./screens/PlayerCreationScreen.jsx";
import StatsticScreen from "./screens/StatisticScreen.jsx";
import AccountSettings from './screens/AccountSettings.jsx';
import { connect, sendMsg } from "./api";

function App() {
    const navigate = useNavigate();
    const [playerCount, setPlayerCount] = useState(2);
    const [players, setPlayers] = useState();
    const [settings, setSettings] = useState({
        startMoney: 2000,
        roundMoney: 200,
        boardSize: 40,
        vacationCash: true,
        auction: true,
        noRentInPrison: true,
        evenBuildings: true,
    });

    useEffect(() => {
        connect(() => {
        });
    }, []);

    function startGame() {
        console.log("Start game:", { playerCount, settings });
        sendMsg("hello");
    }

    return (
        <div className="app">
            <Routes>
                <Route path="/" element={<StartScreen setScreen={(screen) => navigate("/" + screen)} />} />
                <Route path="/players" element={<PlayerScreen playerCount={playerCount} setPlayerCount={setPlayerCount} next={() => navigate("/gamesettings")} back={() => navigate("/")} />} />
                <Route path="/settings" element={<AccountSettings settings={settings} setSettings={setSettings} back={() => navigate("/")} />} />
                <Route path="/join" element={<JoinScreen back={() => navigate("/")} />} />
                <Route path="/player-settings" element={<PlayerCreationScreen playerCount={playerCount} players={players} setPlayers={setPlayers} next={() => navigate("/game/67")} back={() => navigate("/players")} startMoney={settings.startMoney}/>} />
                <Route path="/statistic" element={<StatsticScreen back={() => navigate("/")} />} />
                <Route path="/gamesettings" element={<SettingsScreen settings={settings} setSettings={setSettings} startGame={startGame} next={() => navigate("/player-settings")} back={() => navigate("/players")}/>}/>
            </Routes>
        </div>
    );
}

export default App;