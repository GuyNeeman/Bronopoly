import { logout, send } from "../api/index.js";

export default function StartScreen({ me, go }) {
    return (
        <>
            <h1>Bronopoly</h1>

            {me && (
                <p className="user-bar">
                    Eingeloggt als <b>{me.name}</b>
                    <button className="link" onClick={logout}>Logout</button>
                </p>
            )}

            <p>
                Buy properties, upgrade them, and collect rent.
                The last player with money wins.
            </p>

            <div className="button-row">
                <button onClick={() => go("players")}>Create</button>
                <button onClick={() => go("join")}>Join</button>
                <button onClick={() => go("statistic")}>Statistic</button>
                <button onClick={() => go("settings")}>Settings</button>
                <button onClick={() => send({ action: "ROLL_DICE" })}>Roll Dice</button>
            </div>
        </>
    );
}
