const { spawn } = require('child_process');
const cmd = process.argv.slice(2)

var startTime
const proc = spawn(cmd[0], cmd.slice(1))

proc.stdout.on('data', (data) => {
  if (!data) { return }

  if (!startTime) { startTime = Date.now() }
  const timeDiff = Date.now() - startTime
  process.stdout.write(`${timeDiff}: ${data}`)
});

proc.stderr.on('data', (data) => {
  console.error(`stderr: ${data}`);
});

proc.on('close', (code) => {
  console.log(`child process exited with code ${code}`);
});
