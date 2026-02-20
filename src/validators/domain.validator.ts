import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface,
} from 'class-validator'
import dns from 'dns/promises'

@ValidatorConstraint({ name: 'IsDomain', async: true })
export class IsDomainValidator implements ValidatorConstraintInterface {
  validate(value: string): Promise<boolean> | boolean {
    return domainExists(value)
  }

  defaultMessage(validationArguments?: ValidationArguments): string {
    return `Domain ${validationArguments?.value} does not exist`
  }
}

async function domainExists(domain: string) {
  try {
    const address = await dns.resolve(domain)
    return address.length > 0
  } catch (e) {
    return false
  }
}
