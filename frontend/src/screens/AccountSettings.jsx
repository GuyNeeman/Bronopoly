export default function AccountSettings({back}) {
    return (
        <>
            <h2>Your Settings</h2>
            <p>Your default color is:</p>
            <p>Blue</p>
            <button onClick={back}>Back</button>
        </>
    );
}