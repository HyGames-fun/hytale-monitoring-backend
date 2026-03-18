import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface
} from 'class-validator'

@ValidatorConstraint({ name: 'IsTrimValidator', async: false })
export class IsTrimValidator implements ValidatorConstraintInterface {
  validate(value: string): Promise<boolean> | boolean {
    return value.trim() === value
  }
  defaultMessage?(validationArguments?: ValidationArguments): string {
    return `${validationArguments?.property} must be trimmed`
  }
}
