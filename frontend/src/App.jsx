import { useEffect, useState } from "react";
import { Routes, Route, useNavigate } from "react-router-dom";
import "./App.css";
import { connect, login, MsgType } from "./api";
import LoginScreen from "./screens/LoginScreen.jsx";
import StartScreen from "./screens/StartScreen.jsx";
import PlayerScreen from "./screens/PlayerScreen.jsx";
import SettingsScreen from "./screens/SettingsScreen.jsx";
import JoinScreen from "./screens/JoinScreen.jsx";
import PlayerCreationScreen from "./screens/PlayerCreationScreen.jsx";
import StatisticScreen from "./screens/StatisticScreen.jsx";
import AccountSettings from "./screens/AccountSettings.jsx";

const DEFAULT_SETTINGS = {
    startMoney: 2000,
    roundMoney: 200,
    boardSize: 40,
    vacationCash: true,
    auction: true,
    noRentInPrison: true,
    evenBuildings: true,
};

export default function App() {
    const navigate = useNavigate();
    const back = () => navigate("/");

    // status: "connecting" -> "login" (no/invalid token) -> "ready" (welcome received)
    const [auth, setAuth] = useState({ status: "connecting", me: null, error: null, pending: false });

    const [playerCount, setPlayerCount] = useState(2);
    const [players, setPlayers] = useState();
    const [messages, setMessages] = useState([]);
    const [settings, setSettings] = useState(DEFAULT_SETTINGS);

    useEffect(() => {
        connect(msg => {
            console.log("Received from server:", msg);
            if (msg.type === MsgType.WELCOME) {
                setAuth({ status: "ready", me: msg.body, error: null, pending: false });
            } else if (msg.type === MsgType.LOGIN_FAILED) {
                setAuth({ status: "login", me: null, error: msg.body ?? null, pending: false });
            } else {
                setMessages(prev => [...prev, msg]);
            }
        });
    }, []);

    function handleLogin(name) {
        setAuth(a => ({ ...a, pending: true, error: null }));
        login(name);
    }

    if (auth.status === "connecting") {
        return <div className="app"><p>Verbinde mit Server…</p></div>;
    }

    if (auth.status === "login") {
        return <LoginScreen onLogin={handleLogin} pending={auth.pending} error={auth.error} />;
    }

    return (
        <div className="app">
            <Routes>
                <Route path="/" element={<StartScreen me={auth.me} go={screen => navigate("/" + screen)} />} />
                <Route path="/players" element={
                    <PlayerScreen playerCount={playerCount} setPlayerCount={setPlayerCount}
                        next={() => navigate("/gamesettings")} back={back} />
                } />
                <Route path="/gamesettings" element={
                    <SettingsScreen settings={settings} setSettings={setSettings}
                        next={() => navigate("/player-settings")} back={() => navigate("/players")} />
                } />
                <Route path="/player-settings" element={
                    <PlayerCreationScreen playerCount={playerCount} players={players} setPlayers={setPlayers}
                        startMoney={settings.startMoney}
                        next={() => navigate("/game/67")} back={() => navigate("/gamesettings")} />
                } />
                <Route path="/join" element={<JoinScreen back={back} messages={messages} />} />
                <Route path="/statistic" element={<StatisticScreen back={back} />} />
                <Route path="/settings" element={<AccountSettings back={back} />} />
            </Routes>
        </div>
    );
}
