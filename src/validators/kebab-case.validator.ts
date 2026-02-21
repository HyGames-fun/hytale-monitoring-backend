import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface,
} from 'class-validator'

@ValidatorConstraint({ name: 'IsKebabCase', async: false })
export class IsKebabCaseValidator implements ValidatorConstraintInterface {
  validate(value: string): Promise<boolean> | boolean {
    return /^[a-z0-9]+(-[a-z0-9]+)*$/.test(value)
  }

  defaultMessage(validationArguments?: ValidationArguments): string {
    return `${validationArguments?.value} is not kebab-case`
  }
}
