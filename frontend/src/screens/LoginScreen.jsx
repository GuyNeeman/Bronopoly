import { useState } from "react";
import "./LoginScreen.css";

const MAX_NAME = 20; // same limit as cleanName on the server

export default function LoginScreen({ onLogin, pending, error }) {
    const [name, setName] = useState("");
    const trimmed = name.trim();
    const canSubmit = trimmed.length > 0 && !pending;

    function handleSubmit(e) {
        e.preventDefault(); // stop the form from reloading the page
        if (canSubmit) onLogin(trimmed);
    }

    return (
        <div className="login-page">
            <form className="login-card" onSubmit={handleSubmit}>
                <div className="login-strip" />

                <div className="login-body">
                    <p className="login-kicker">Willkommen bei</p>
                    <h1 className="login-title">Bronopoly</h1>

                    <label className="login-label" htmlFor="player-name">
                        Wie sollen dich die anderen nennen?
                    </label>
                    <input
                        id="player-name"
                        className="login-input"
                        value={name}
                        onChange={e => setName(e.target.value)}
                        maxLength={MAX_NAME}
                        placeholder="Dein Name"
                        autoComplete="nickname"
                        autoFocus
                        disabled={pending}
                    />
                    <div className="login-counter">{name.length}/{MAX_NAME}</div>

                    {error && <p className="login-error">{error}</p>}

                    <button className="login-button" type="submit" disabled={!canSubmit}>
                        {pending ? "Verbinde…" : "Spielen"}
                    </button>

                    <p className="login-hint">Kein Passwort nötig. Dein Browser merkt sich dich.</p>
                </div>
            </form>
        </div>
    );
}
