
import * as child_process from 'child_process';

const exec = (command) => {
  return new Promise((resolve, reject) => {
    const p = child_process.spawn(command, []);
    let isResolved = false
    p.stderr.on('data', (data) => {
      console.error(`stderr: ${data}`);
      if (isResolved) return
      isResolved = true
      resolve()
    });
    
    p.on('close', (code) => {
      console.log(`child process exited with code ${code}`);
    }); 
  })
};

const findPid = (processName) => {
  return new Promise((resolve, reject) => {
    if (process.platform === 'win32') {
      child_process.exec(`wmic process where caption="${processName}" get ProcessId`, (err, stdout, stderr) => {
        if (err) {
          reject(err)
        } else {
          const lines = stdout.replace(/\r\r\n/gm, "\r\n").split("\r\n").filter(i => 0 < i.trim().length);
          if (lines.length <= 1) {
            reject(new Error('No process found'));
          } else {
            resolve(lines[1]);
          }
        }
      });
    } else if (process.platform === 'darwin') {
      child_process.exec(`ps -ef | grep ${processName} | grep -v grep | awk '{print $2}'`, (err, stdout, stderr) => {
        if (err) {
          reject(err)
        } else {
          resolve(stdout);
        }
      });
    } else {
      reject(new Error('Unsupported platform'));
    }
  })
};

const killProcess = (pidOrName) => {
  return new Promise((resolve, reject) => {
    if (typeof pidOrName === 'number') {
      resolve(process.kill(pidOrName, 'SIGINT'));
    } else {
      findPid(pidOrName).then(pid => {
        if (pid) {
          resolve(process.kill(pid, 'SIGINT'));
        } else {
          resolve(false);
        }
      }).catch(err => reject(err));
    }
  });
};

(async () => {
  const result = await exec('/Users/longzy/Desktop/gomedia');
  console.log(result)
  console.log('执行成功...')
})();

export { exec, findPid, killProcess };