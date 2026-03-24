import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface
} from 'class-validator'

ValidatorConstraint({ name: 'IsLatin', async: false })
export class IsLatinValidator implements ValidatorConstraintInterface {
  validate(value: any): Promise<boolean> | boolean {
    return typeof value === 'string' && /^[A-Za-z]+$/.test(value)
  }

  defaultMessage(args?: ValidationArguments): string {
    return `${args?.property} must be only latin`
  }
}
