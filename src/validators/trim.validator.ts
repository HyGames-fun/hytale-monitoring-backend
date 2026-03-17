import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface
} from 'class-validator'

@ValidatorConstraint({ name: 'TrimValidator', async: false })
export class TrimValidator implements ValidatorConstraintInterface {
  validate(value: string): Promise<boolean> | boolean {
    return value.trim() === value
  }
  defaultMessage?(validationArguments?: ValidationArguments): string {
    return `${validationArguments?.property} must be trimmed`
  }
}
