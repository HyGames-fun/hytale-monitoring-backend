import { registerDecorator, ValidationOptions } from 'class-validator'
import { AtLeastOneValidator } from '../validators/at-least-one.validator'

export function AtLeastOne(
  properties: string[],
  validationOptions?: ValidationOptions,
) {
  return function (object: object, propertyName: string) {
    registerDecorator({
      target: object.constructor,
      propertyName,
      options: validationOptions,
      constraints: properties,
      validator: AtLeastOneValidator,
    })
  }
}