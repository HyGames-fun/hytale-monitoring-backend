import {
  ValidationArguments,
  ValidatorConstraint,
  ValidatorConstraintInterface
} from 'class-validator'

@ValidatorConstraint({ async: false })
export class AtLeastOneValidator implements ValidatorConstraintInterface {
  validate(_: any, args: ValidationArguments) {
    const obj = args.object as Record<string, unknown>
    const properties: string[] = args.constraints as string[]

    return properties.some((prop) => {
      const value = obj[prop]
      return (
        value !== null &&
        value !== undefined &&
        (typeof value !== 'string' || value.trim() !== '')
      )
    })
  }

  defaultMessage(args: ValidationArguments) {
    return `At least one of the following properties must be provided: ${args.constraints.join(
      ', '
    )}`
  }
}
