export default function JoinScreen({back}) {
    return (
        <>
            <h2>Join a game</h2>
            <p>Put the code you received in the following input:</p>
            <input type="text"/>
            <button onClick={back}>Back</button>
        </>
    );
}