import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface
} from 'class-validator'
import dns from 'dns/promises'

@ValidatorConstraint({ name: 'IsDomain', async: true })
export class IsDomainValidator implements ValidatorConstraintInterface {
  validate(value: string): Promise<boolean> | boolean {
    if (!value) return false

    if (value.includes(':')) {
      const [domain, port] = value.split(':', 2)
      if (!domain && !port) return false

      const portNum = +port
      if (isNaN(portNum) || portNum < 1 || portNum > 65535) {
        return false
      }

      return domainExists(domain)
    }

    return domainExists(value)
  }

  defaultMessage(validationArguments?: ValidationArguments): string {
    return `domain ${validationArguments?.value} does not exist or incorrect format`
  }
}

async function domainExists(domain: string) {
  try {
    await dns.lookup(domain)
    return true
  } catch (e) {
    return false
  }
}
