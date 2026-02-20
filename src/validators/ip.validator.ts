import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface,
} from 'class-validator'
import net from 'net'

@ValidatorConstraint({ name: 'IsIp', async: true })
export class IsIpValidator implements ValidatorConstraintInterface {
  validate(value: string): Promise<boolean> | boolean {
    const [host, port] = value.split(':')

    if (!host || !port || isNaN(+port)) return false

    return checkPort(host, port)
  }

  defaultMessage(validationArguments?: ValidationArguments): string {
    return `Host or port is not reachable in ${validationArguments?.value}`
  }
}

async function checkPort(
  host: string,
  port: string,
  timeout: number = 3000,
): Promise<boolean> {
  return new Promise((resolve) => {
    const socket = new net.Socket()

    socket.setTimeout(timeout)

    socket.on('connect', () => {
      resolve(true)
    })

    socket.on('timeout', () => {
      resolve(false)
    })

    socket.on('error', () => {
      resolve(false)
    })

    socket.connect(+port, host)
  })
}
