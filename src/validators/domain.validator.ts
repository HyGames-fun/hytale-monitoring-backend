import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface,
} from 'class-validator'
import dns from 'dns/promises'

@ValidatorConstraint({ name: 'IsDomain', async: true })
export class IsDomainValidator implements ValidatorConstraintInterface {
  validate(value: string): Promise<boolean> | boolean {
    if (!value) return false

    return domainExists(value)
  }

  defaultMessage(validationArguments?: ValidationArguments): string {
    return `domain ${validationArguments?.value} does not exist`
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
